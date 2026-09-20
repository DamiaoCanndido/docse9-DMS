# Remediação de Vulnerabilidades de Segurança: Upload e Gestão de Anexos (Cloudflare R2) — Spec

**Status:** Approved  
**Author:** @security / @pm  
**Date:** 2026-09-20  
**Related:** `.agents/workflows/startcycle.md`, `docs/specs/document-file-upload-cloudflare-r2.md`, `backend/internal/service/document_service.go`, `backend/pkg/storage/storage.go`, `backend/pkg/storage/pdf_validator.go`, `backend/internal/domain/document.go`, `backend/cmd/api/main.go`

---

## Problem Statement

Durante a auditoria de segurança da funcionalidade de anexos e upload direto no Cloudflare R2, foram identificadas vulnerabilidades críticas de segurança que comprometem o isolamento multi-tenant, a integridade dos dados e a disponibilidade da API. Especificamente:
1. O endpoint de confirmação de upload (`POST /documents/:id/confirm-upload`) aceita chaves de objeto arbitrárias (`fileKey`) sem validação de tenant/documento, permitindo sequestro e exfiltração de documentos confidenciais de outros municípios (IDOR/BOLA) e exclusão arbitrária de objetos no bucket R2.
2. O endpoint de atualização (`PATCH /documents/:id`) aceita alteração direta de `fileKey` via Mass Assignment, ignorando o fluxo seguro de upload e as validações de PDF/OCR.
3. A inspeção de OCR no backend (`storage.ValidatePDFOCR`) utiliza `io.ReadAll` sem teto de bytes em instâncias com memória limitada no Google Cloud Run, expondo o serviço a Denial of Service (OOM Crash).
4. A geração de URLs pré-assinadas não impõe limite de tamanho no S3 (`ContentLength`), permitindo uploads excessivos diretamente no R2.

Esta especificação define as correções arquiteturais e de código necessárias para mitigar integralmente essas vulnerabilidades antes da liberação do sistema para múltiplos municípios em produção.

---

## Goals

- Eliminar vulnerabilidades de IDOR e exfiltração de documentos cross-tenant no fluxo de upload e confirmação.
- Impor validação estrita de prefixo multi-tenant no `ConfirmUpload`: a chave deve obrigatoriamente respeitar a estrutura `tenants/{municipalityId}/{documentType}/{year}/{documentId}/{uuid}.pdf`.
- Remover a manipulação direta de `fileKey` no endpoint `PATCH /documents/:id` (extinguir vetor de Mass Assignment).
- Proteger rotinas de exclusão (`DeleteObject`) para que nunca excluam objetos fora do escopo validado do documento em tratamento.
- Blindar o backend contra esgotamento de memória (OOM / DoS) impondo teto máximo de 25MB via `io.LimitReader` e validação prévia de `ContentLength` no leitor S3.
- Incluir `ContentLength` assinado na geração de Presigned PUT URLs no Cloudflare R2.
- Adicionar proteção com `recover()`, teto de páginas e verificação de cancelamento de contexto no leitor de PDF/OCR (`ValidatePDFText`).
- Garantir `ResponseContentType: application/pdf` nas URLs pré-assinadas de visualização/download para prevenir MIME Sniffing e Stored XSS.
- Atualizar a suíte de testes unitários e de integração no backend cobrindo todas as novas regras de proteção.

---

## Non-Goals

- Não alterar a tecnologia de armazenamento (permanece Cloudflare R2 compatível com S3).
- Não alterar a regra de negócio central de exigência de camada de texto/OCR pesquisável em PDFs.
- Não introduzir dependências externas adicionais no backend Go além das bibliotecas já utilizadas (`aws-sdk-go-v2`, `ledongthuc/pdf`).

---

## Background / Context

O Docseq DMS opera sob arquitetura multi-tenant onde múltiplos municípios compartilham o mesmo banco de dados e bucket R2, segregados logicamente por `MunicipalityID`. Em conformidade com a LGPD (Lei nº 13.709/2018) e as diretrizes de integridade pública municipal, o acesso não autorizado a documentos administrativos entre entes federativos distintos configura violação gravíssima de privacidade e governança.

O backend roda em contêineres serverless no Google Cloud Run, caracterizados por memória estrita (512MB–1GB). Qualquer consumo não delimitado de buffers em memória RAM ameaça a continuidade operacional de todos os municípios atendidos.

---

## Vulnerabilities & Threat Model (STRIDE)

| ID | Vulnerabilidade | STRIDE | Severidade | CVSS v3.1 |
| :---: | :--- | :---: | :---: | :---: |
| **VULN-01** | IDOR e exfiltração cross-tenant de PDFs via `ConfirmUpload` | Elevation of Privilege / Tampering | **Crítica** | 8.8 |
| **VULN-02** | Oráculo de exclusão arbitrária de objetos no Cloudflare R2 | Tampering / Repudiation | **Crítica** | 8.6 |
| **VULN-03** | Mass Assignment de `fileKey` no `PATCH /documents/:id` | Elevation of Privilege | **Crítica** | 8.1 |
| **VULN-04** | DoS / OOM Crash via `io.ReadAll` ilimitado em `ValidatePDFOCR` | Denial of Service | **Alta** | 7.5 |
| **VULN-05** | Falta de imposição de tamanho máximo na Presigned Upload URL | Denial of Service / Wallet | **Alta** | 6.5 |
| **VULN-06** | Parser Crash e CPU Pegging em `ValidatePDFText` | Denial of Service | **Média** | 5.3 |
| **VULN-07** | Ausência de `ResponseContentType: application/pdf` em download inline | Information Disclosure / XSS | **Média** | 4.7 |

---

## Requirements

### Functional

1. **Validação Estrita de Chave em `ConfirmUpload` (VULN-01 e VULN-02):**
   - No método `documentService.ConfirmUpload`:
     - Calcular o prefixo canônico esperado: `expectedPrefix := fmt.Sprintf("tenants/%s/%s/%d/%s/", doc.MunicipalityID, doc.Type, doc.CreatedAt.Year(), doc.ID)`.
     - Validar que `strings.HasPrefix(cleanKey, expectedPrefix)`.
     - Validar que `strings.HasSuffix(strings.ToLower(cleanKey), ".pdf")`.
     - Se a chave não respeitar o prefixo exato, rejeitar a requisição com erro `400 Bad Request` ("fileKey inválido ou não pertence a este documento") **sem executar nenhuma exclusão no R2**.
   - Garantir que a rotina `_ = s.storageSvc.DeleteObject(ctx, cleanKey)` só seja chamada caso o arquivo pertença comprovadamente ao prefixo do documento validado.

2. **Remoção de Mass Assignment em `UpdateDocumentInput` (VULN-03):**
   - Remover o campo `FileKey *string` do struct `domain.UpdateDocumentInput`.
   - Remover o bloco de atribuição de `input.FileKey` no método `documentService.Update`.
   - Modificações ou anexações de arquivos passam a ser permitidas exclusivamente pelos endpoints `/upload-url` e `/confirm-upload`.

3. **Imposição de Tamanho na Geração de Upload URL (VULN-05):**
   - Atualizar a assinatura da interface `StorageService`:
     ```go
     GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, contentLength int64, expiresIn time.Duration) (string, error)
     ```
   - Em `pkg/storage/storage.go`, incluir `ContentLength: aws.Int64(contentLength)` no `s3.PutObjectInput`.
   - Passar `input.FileSize` a partir de `documentService.GenerateUploadURL`.

4. **Tratamento de Exibição Inline Segura (VULN-07):**
   - Em `storage.GeneratePresignedDownloadURL`, definir explicitamente `ResponseContentType = aws.String("application/pdf")` no `s3.GetObjectInput` para coibir MIME Sniffing no navegador.

### Non-Functional & Resiliência

1. **Proteção Contra Esgotamento de Memória (VULN-04):**
   - Em `storage.ValidatePDFOCR`:
     - Constante `maxAllowedBytes = 25 * 1024 * 1024` (25 MB).
     - Se `out.ContentLength != nil && *out.ContentLength > maxAllowedBytes`, abortar com erro imediato sem ler o corpo.
     - Envolver a leitura com `io.LimitReader(out.Body, maxAllowedBytes+1)`.
     - Se o número de bytes lidos exceder `maxAllowedBytes`, descartar o buffer e retornar erro de tamanho excedido.

2. **Mitigação de Pânico e DoS de CPU no Parser PDF (VULN-06):**
   - Em `pkg/storage/pdf_validator.go` (`ValidatePDFText`):
     - Adicionar bloco `recover()` com captura de pânico para evitar derrubar o processo em arquivos corrompidos.
     - Limitar o loop de inspeção de páginas a no máximo 50 páginas (`maxPagesToCheck = min(numPages, 50)`).
     - Aceitar `context.Context` (ou checar `ctx.Err()`) caso cancelado pelo cliente.

---

## Acceptance Criteria

- [ ] `POST /documents/:id/confirm-upload` rejeita com erro 400 qualquer `fileKey` que não comece com o prefixo exato do município, tipo, ano e ID do documento.
- [ ] `POST /documents/:id/confirm-upload` não exclui nenhum arquivo do R2 se o `fileKey` informado for de outro município ou caminho arbitrário.
- [ ] `PATCH /documents/:id` não aceita mais o campo `fileKey` no payload JSON.
- [ ] `GeneratePresignedUploadURL` assina o cabeçalho `Content-Length`, impedindo uploads maiores do que o valor validado.
- [ ] `ValidatePDFOCR` rejeita arquivos maiores que 25MB sem estourar a memória RAM da aplicação.
- [ ] `ValidatePDFText` trata pânicos internamente via `recover()`, retornando erro amigável.
- [ ] `GeneratePresignedDownloadURL` força o cabeçalho de resposta `Content-Type: application/pdf`.
- [ ] Todos os testes unitários do backend (`go test ./...`) e frontend (`npm test`) passam com 100% de sucesso.
- [ ] Novos testes unitários cobrem:
  - Rejeição de chave cross-tenant em `ConfirmUpload`.
  - Rejeição de chave com extensão ou prefixo inválido.
  - Tentativa de Mass Assignment via `Update`.
  - Proteção de limite de leitura de bytes no storage.

---

## Success Metrics

- 0 vulnerabilidades críticas ou altas remanescentes na auditoria de segurança de upload.
- 100% de isolamento verificado entre tenants em ambiente multi-município.
- Estabilidade operacional garantida sem incidentes de OOM no Cloud Run sob carga de documentos.
