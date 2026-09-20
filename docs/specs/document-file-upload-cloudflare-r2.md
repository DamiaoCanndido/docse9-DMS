# Upload e Gestão de Anexos com Cloudflare R2 — Spec

**Status:** Approved  
**Author:** @pm  
**Date:** 2026-09-20  
**Related:** `backend/internal/domain/document.go`, `backend/internal/domain/permission.go`, `backend/internal/handler/document_handler.go`, `backend/cmd/api/main.go`, `frontend/src/components/documents/DocumentFormDialog.tsx`, `frontend/src/components/documents/DocumentTable.tsx`

---

## Problem Statement
Atualmente, o Docseq DMS gerencia apenas os metadados dos documentos oficiais (número de ordem, descrição, tipo e valores contratuais), mantendo o campo `fileKey` em branco. Os servidores e gestores municipais necessitam anexar o arquivo digital oficial de cada ato administrativo para conferir validade jurídica, transparência pública e perenidade ao acervo municipal. Para atender às diretrizes da Lei de Acesso à Informação (LAI - Lei nº 12.527/2011) e do Marco Legal do Governo Digital (Lei nº 14.129/2021), é mandatório que cada documento possua estritamente um único arquivo em formato PDF e que este contenha obrigatoriamente camada de texto pesquisável (OCR ativo), sendo sumariamente rejeitados arquivos compostos unicamente por imagens escaneadas sem texto.

---

## Goals
- Integrar o serviço de armazenamento de objetos **Cloudflare R2** compatível com a API S3 no backend Go com custo zero de tráfego de saída (*zero egress fees*).
- Restringir o anexo a **estritamente 1 (um) arquivo por documento**.
- Restringir o formato de anexo exclusivamente ao tipo MIME **`application/pdf` (`.pdf`)**.
- Implementar **validação obrigatória de camada de texto (OCR)**: rejeitar PDFs escaneados puramente como imagens que não possuam texto pesquisável/extraível.
- Disponibilizar upload direto via **Presigned URLs** (URLs pré-assinadas) para envio do frontend diretamente ao Cloudflare R2, respeitando o limite de 1MB do payload da API no Cloud Run.
- Permitir geração de URLs pré-assinadas seguras e temporárias para visualização inline (preview) e download do PDF anexado.
- Reutilizar a matriz granular de permissões existente: usuários com nível `WRITE` ou `DELETE` no tipo de documento (e moderadores `MOD`) podem anexar/substituir o PDF; usuários com nível `READ` podem visualizar/baixar.
- Atualizar a interface web (`DocumentFormDialog` e `DocumentTable`) para exibir campos de anexo em PDF, validação visual de OCR em tempo real, barra de progresso e botão de visualização/download.

---

## Non-Goals
- Não executar o processo de conversão/extração de OCR pesado no servidor (o arquivo deve obrigatoriamente **já conter a camada de OCR prévia** proveniente do scanner ou software gerador; caso contrário, será recusado na validação).
- Não permitir múltiplos anexos para um mesmo documento (relação estrita de 1 documento para 1 PDF oficial).
- Não aceitar formatos alternativos de arquivo (imagens PNG/JPEG, arquivos Word/DOCX, planilhas ou compactados não serão permitidos).
- Não expor o bucket do Cloudflare R2 como público (todos os acessos aos PDFs devem ser estritamente controlados por URLs assinadas com tempo de expiração).
- Não criar um novo nível de permissão isolado no banco de dados para upload (a governança segue a permissão de escrita/leitura do tipo de documento).

---

## Background / Context
- **Conformidade Legal e Acessibilidade:** Documentos públicos digitalizados sem camada de texto (OCR) violam preceitos de acessibilidade (leitores de tela para pessoas com deficiência visual) e impedem indexação e busca textual por órgãos de controle (Tribunais de Contas) e cidadãos. Por essa razão, a exigência de OCR por padrão é uma diretriz inegociável de governança.
- **Arquitetura Serverless & Limite de Payload:** O backend Go roda no Google Cloud Run e possui um middleware de segurança limitando o corpo de requisições HTTP em 1MB (`middleware.MaxBodySizeMiddleware(1 << 20)` em `backend/cmd/api/main.go`). Para acomodar PDFs municipais (frequentemente entre 2MB e 25MB), o envio do binário deve ser feito via upload direto com Presigned URL do Cloudflare R2.
- **Isolamento Multi-Tenant:** Cada município (`MunicipalityID`) opera de forma segregada. No bucket R2, a chave do arquivo (*Object Key*) deve obrigatoriamente incluir o identificador do município para garantir isolamento e auditabilidade.

---

## User Stories
1. **Como Moderador (`MOD`) ou Operador Municipal (`COMMON` com permissão WRITE):**
   Quero anexar o PDF oficial com OCR ao cadastrar ou editar um documento, para que o documento oficial fique preservado e acessível no arquivo digital.
2. **Como Moderador ou Operador:**
   Quero ser avisado imediatamente se o PDF selecionado não tiver camada de OCR ou se não for um arquivo PDF, para que eu possa providenciar a versão correta antes de salvar o documento.
3. **Como Operador Municipal (`COMMON` com permissão READ):**
   Quero visualizar ou baixar o PDF oficial anexado a um documento do meu município, conferindo o teor original na íntegra.
4. **Como Gestor do Sistema:**
   Quero garantir que nenhum documento acumule múltiplos arquivos soltos e que todo arquivo anexado seja pesquisável e com tráfego livre de cobrança de egress.

---

## Requirements

### Functional

#### 1. Regras de Negócio e Validação de Arquivos
- **Cardinalidade:** Estritamente **1 único arquivo PDF por documento**. Se um novo PDF for enviado em uma edição, o arquivo anterior é substituído e removido do Cloudflare R2.
- **Tipo MIME Permitido:** Exclusivamente `application/pdf` (extensão `.pdf`). Qualquer outro tipo MIME deve ser rejeitado.
- **Tamanho Máximo:** Até **25 MB** por PDF (configurável via variável de ambiente `MAX_FILE_SIZE_MB=25`).
- **Validação Obrigatória de Camada OCR (Texto Pesquisável):**
  - O PDF anexado deve obrigatoriamente conter texto digital pesquisável (vetorial ou camada de OCR invisível gerada na digitalização).
  - PDFs puramente baseados em imagens rasterizadas (scans planos sem texto extraível) devem ser **rejeitados com mensagem explicativa**:
    *"O arquivo PDF selecionado não possui camada de texto pesquisável (OCR). Por favor, realize a digitalização com OCR habilitado antes de anexar."*
  - **Estratégia de Validação em Camadas:**
    1. **Client-Side (Frontend):** Utilizar `pdfjs-dist` (ou leitor local equivalente) para inspecionar os blocos de texto das páginas no navegador antes do upload. Se nenhuma página contiver texto extraível ou a contagem de texto for nula em páginas rasterizadas, o upload é abortado preventivamente na interface.
    2. **Server-Side (Backend):** No endpoint de confirmação/inspeção do anexo (`POST /documents/:id/confirm-upload` ou no `PATCH /documents/:id`), o backend inspeciona o cabeçalho e streams de texto/fontes do PDF (usando biblioteca Go como `github.com/ledongthuc/pdf` ou `github.com/pdfcpu/pdfcpu`). Se confirmada a ausência de texto, o backend descarta o objeto do R2 e retorna `422 Unprocessable Entity`.

#### 2. Backend (Go / Gin / S3 SDK)
- **Pacote de Storage (`pkg/storage` ou `internal/storage`):**
  - Implementar client S3 utilizando o SDK oficial da AWS para Go (`github.com/aws/aws-sdk-go-v2/service/s3`).
  - Variáveis de ambiente requeridas:
    - `R2_ACCOUNT_ID`: ID da conta Cloudflare.
    - `R2_ACCESS_KEY_ID`: Chave de acesso do R2.
    - `R2_SECRET_ACCESS_KEY`: Chave secreta do R2.
    - `R2_BUCKET_NAME`: Nome do bucket (ex: `docseq-documents`).
    - Endpoint construído: `https://<R2_ACCOUNT_ID>.r2.cloudflarestorage.com`.
- **Geração de Presigned URL para Upload (`POST /documents/:id/upload-url`):**
  - Entrada: `fileName` (string), `fileSize` (int64), `contentType` (string).
  - Validações:
    - Rejeitar se `contentType != "application/pdf"`.
    - Rejeitar se `fileSize > 25MB`.
    - Permissões: `MOD` ou `COMMON` com nível `WRITE`/`DELETE` no tipo do documento. O documento deve pertencer ao mesmo município do usuário.
  - Estrutura da Chave de Objeto no R2:
    `tenants/{municipalityId}/{documentType}/{year}/{documentId}/{uuid}.pdf`
  - Retorno:
    ```json
    {
      "uploadUrl": "https://<account>.r2.cloudflarestorage.com/...",
      "fileKey": "tenants/...",
      "expiresInSeconds": 600
    }
    ```
- **Confirmação e Validação de OCR (`POST /documents/:id/confirm-upload`):**
  - Entrada: `fileKey` (string).
  - O backend verifica o arquivo no R2, valida que é um PDF válido e confirma a presença de camada de texto (OCR).
  - Se válido: vincula o `fileKey` ao documento no banco de dados e remove qualquer arquivo antigo anterior do mesmo documento.
  - Se inválido (sem OCR): exclui o objeto do R2 e retorna `422 Unprocessable Entity` com mensagem de erro detalhada.
- **Geração de Presigned URL para Visualização / Download (`GET /documents/:id/file-url`):**
  - Validação de permissões: Usuário com nível `READ`, `WRITE` ou `DELETE` no tipo de documento.
  - Se o documento não possuir `fileKey`, retornar `404 Not Found` ("documento não possui arquivo anexo").
  - Gera URL assinada de leitura (`s3.PresignGetObject`) com expiração de 15 minutos (900s).
  - Suportar query param opcional `?download=true` para forçar download direto via cabeçalho `Content-Disposition: attachment; filename="<descricao_ou_numero>.pdf"`. Caso contrário, gera para visualização em aba (`inline`).
- **Exclusão de Arquivo:**
  - Quando um documento for excluído definitivamente (`HardDelete`), o arquivo correspondente no Cloudflare R2 é removido.

#### 3. Frontend (Next.js 16 / React 19 / Tailwind / Shadcn UI)
- **Modal de Formulário (`DocumentFormDialog.tsx`):**
  - Campo exclusivo de upload de arquivo PDF (Dropzone / Input File aceitando apenas `.pdf` / `application/pdf`).
  - Exibição de alerta informativo: *"Apenas arquivos em formato PDF com OCR (camada de texto pesquisável) são aceitos."*
  - Pré-validação client-side:
    - Se o usuário tentar arrastar um arquivo que não seja `.pdf`, exibir aviso imediato via `sonner`.
    - Checagem rápida de OCR: inspecionar se há caracteres extraíveis no PDF. Se não houver, bloquear o envio e exibir mensagem clara.
  - Exibição do estado do anexo: nome do arquivo, tamanho em MB, indicador de OCR verificado e botão para remover/trocar arquivo.
  - Barra de progresso visual durante o upload direto no R2.
- **Tabela de Documentos (`DocumentTable.tsx`):**
  - Ícone/badge indicando a presença de PDF anexado.
  - Ação rápida no menu de contexto: "Visualizar PDF" (abre nova aba com a Presigned URL) e "Baixar PDF".

---

## Non-Functional Requirements
- **Segurança & LGPD:**
  - Nenhuma URL pré-assinada deve possuir expiração superior a 15 minutos.
  - Bucket R2 configurado estritamente como privado (`Public Access Disabled`).
  - O backend nunca emite URLs de arquivos pertencentes a municípios diferentes do usuário solicitante.
- **Performance & Recursos:**
  - O tráfego do arquivo de até 25MB não passa pelo backend no Cloud Run, preservando a memória do container e a largura de banda.
  - A inspeção de OCR no backend deve ocorrer de forma otimizada (lendo apenas metadados/dicionários de fontes e streams de texto sem renderização gráfica de páginas).
- **Usabilidade & Acessibilidade:**
  - Mensagens de erro amigáveis orientando o operador caso o scanner da prefeitura não tenha aplicado OCR.

---

## Out of Scope
- Aplicação automática de OCR no servidor para PDFs escaneados sem texto (conversão via Tesseract/Cloud Vision está fora de escopo).
- Upload de múltiplos arquivos para o mesmo documento.
- Suporte a extensões `.docx`, `.png`, `.jpeg` ou outros formatos.

---

## Acceptance Criteria
- [ ] O sistema rejeita imediatamente qualquer arquivo com extensão ou MIME type diferente de `application/pdf`.
- [ ] O sistema valida a presença de camada de texto (OCR) no PDF; arquivos compostos unicamente de imagens escaneadas sem OCR são rejeitados com status `422 Unprocessable Entity` e mensagem orientativa.
- [ ] Documentos aceitam estritamente 1 único arquivo PDF associado ao campo `fileKey`.
- [ ] A substituição de um anexo remove o arquivo anterior do Cloudflare R2, evitando arquivos órfãos.
- [ ] A deleção definitiva (`HardDelete`) do documento remove o PDF correspondente no Cloudflare R2.
- [ ] `POST /documents/:id/upload-url` rejeita requisições de operadores sem permissão `WRITE`/`DELETE` no tipo de documento ou de outro município com `403 Forbidden`.
- [ ] `GET /documents/:id/file-url` gera link de visualização/download temporário apenas para quem tem permissão `READ` ou superior.
- [ ] Upload é executado diretamente do navegador para o Cloudflare R2 via `PUT` sem estourar o limite de 1MB da API.
- [ ] Testes automatizados no backend cobrem:
  - Validação de MIME type exclusivo (`application/pdf`).
  - Rejeição de PDF sem OCR.
  - Aprovação de PDF com texto/OCR válido.
  - Controle de acesso por tipo de documento e município.

---

## Open Questions & Risks
- **Configuração de CORS no Cloudflare R2:** O bucket R2 precisa autorizar os métodos `PUT` e `GET` com o cabeçalho `Content-Type: application/pdf` para as origens do frontend (`http://localhost:3000` em dev e domínio Netlify em prod).
- **Variação na Qualidade de OCR:** PDFs com texto incompleto (ex: apenas a folha de rosto com texto e anexos ilegíveis).
  - *Critério de Aceite:* Para ser aprovado, o documento deve conter ao menos 50 caracteres alfanuméricos de texto extraível distribuídos pelo documento.

---

## Success Metrics
- 100% dos documentos oficiais cadastrados com anexos válidos em PDF pesquisável.
- Zero arquivos de formato inválido (não-PDF) no repositório de armazenamento.
- Zero ocorrências de arquivos órfãos após substituições e exclusões definitivas.
- Zero custo de egress no Cloudflare R2 para os municípios.

---

## Rollout Plan
1. **Fase 1 (Infraestrutura):** Criar bucket no Cloudflare R2, gerar API Token e configurar política de CORS para `application/pdf`.
2. **Fase 2 (Backend):**
   - Configurar cliente S3/R2 no Go.
   - Implementar validador de PDF e leitor de camada de texto (OCR).
   - Implementar rotas `/upload-url`, `/confirm-upload` e `/file-url`.
   - Adicionar rotinas de limpeza no R2 para substituição e `HardDelete`.
3. **Fase 3 (Frontend):**
   - Adicionar pré-checagem de PDF e OCR com `pdfjs-dist` no `DocumentFormDialog`.
   - Implementar barra de progresso no upload direto para o R2.
   - Atualizar `DocumentTable` com ações de visualização/download.
4. **Fase 4 (QA & Homologação):** Validar com PDFs de texto nato, PDFs com OCR aplicado via scanner e PDFs de imagem pura (para garantir a rejeição adequada).
