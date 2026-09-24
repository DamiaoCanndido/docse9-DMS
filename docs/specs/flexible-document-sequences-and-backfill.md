# Gestão de Sequências Numéricas Flexíveis, Marcos Iniciais e Cadastro Retroativo — Spec

**Status:** Approved  
**Author:** @pm  
**Date:** 2026-09-24  
**Related:** `backend/internal/domain/document.go`, `backend/internal/service/document_service.go`, `backend/internal/repository/document_repository.go`, `frontend/src/components/documents/DocumentFormDialog.tsx`, `docs/specs/document-file-upload-cloudflare-r2.md`

---

## Problem Statement

Atualmente, o sistema calcula a numeração de documentos (`order`) sempre a partir de `COALESCE(MAX(order), 0) + 1` e vincula o exercício ao relógio do servidor (`time.Now().Year()`), sem permitir que o usuário defina a data oficial do ato nem o número inicial da série. Quando um município adota o Docseq no meio do ano (já tendo emitido dezenas de atos no papel ou planilha) ou precisa registrar documentos históricos de anos e meses anteriores, o sistema força o início no nº 1 ou insere os atos no ano atual, gerando duplicidade jurídica com documentos físicos e inviabilizando a digitalização de acervos legados.

---

## Goals

1. **Seleção da Data Oficial do Ato na Criação:** Permitir que o operador defina a data e hora oficial do documento no momento do cadastro (com valor padrão = momento atual), particionando a sequência numérica anual com base no ano da data informada.
2. **Configuração de Marco Inicial de Sequência (Offset):** Permitir que o Moderador (`MOD`) configure o número de partida (`initial_order`) para cada tipo de documento e ano (e contínuo perpétuo para Leis), garantindo que municípios que aderirem no meio do ano iniciem a numeração automática a partir do número subsequente ao físico (ex: iniciar Decretos 2026 no nº 85).
3. **Lançamento Manual para Acervo Legado/Físico:** Permitir exclusivamente a Moderadores (`MOD`) informar manualmente o número (`order`) de um documento legado durante a digitalização de acervos físicos ou planilhas.
4. **Validação Estrita de Integridade e Unicidade:** Bloquear duplicidade de números dentro do mesmo município, tipo e ano (ou geral para Leis), com mensagens de erro claras, garantindo que a geração automática subsequente respeite os números manuais já inseridos sem gerar conflitos.
5. **Manutenção da Atomicidade Concorrente:** Preservar a integridade transacional com advisory lock (`pg_advisory_xact_lock`) no cálculo de números concorrentes.

---

## Non-Goals

- Não permitir a alteração arbitrária de numeração após o documento já ter sido criado (a numeração oficial do ato permanece imutável após salva para garantir a fé pública).
- Não suportar numerações com sufixos alfanuméricos ou letras (ex: "85-A"); a numeração oficial segue estritamente o padrão de inteiros sequenciais (`order: integer`).
- Não permitir que usuários com perfil `ADMIN` ou `COMMON` sobrescrevam ou digitem números manuais nem alterem offsets (privilégio estritamente reservado ao perfil `MOD` — Moderador/Gestor Documental).
- Não substituir o banco relacional por geradores de IDs externos; a persistência de sequências permanece integrada ao PostgreSQL.

---

## Background / Context

Na administração pública municipal (Leis Federais nº 4.320/64 e 14.133/21, bem como manuais de redação oficial), atos normativos (Decretos, Portarias, Editais e Contratos) são numerados anualmente (ex: Portaria nº 001/2026, Contrato nº 015/2026), reiniciando o ciclo a cada 1º de janeiro. Já as Leis Ordinárias e Complementares possuem numeração contínua e perpétua que atravessa mandatos e décadas.

Na prática real de implantação de software nas prefeituras:
- A adesão raramente ocorre exatamente no dia 1º de janeiro; na maioria das vezes ocorre no meio do exercício financeiro.
- O município já possui atos emitidos fisicamente que não podem ter seus números repetidos.
- Há equipes dedicadas a digitalizar o histórico de meses anteriores e exercícios anteriores para centralizar as buscas no DMS.

---

## User Stories

1. **Como Moderador (`MOD`)**, quero configurar o marco inicial de numeração de Decretos de 2026 no número 85, para que o primeiro decreto gerado automaticamente no sistema continue a sequência física sem reiniciar do 1.
2. **Como Operador de Gabinete**, quero cadastrar um decreto assinado em 15 de janeiro informando a data correta no formulário, para que o sistema registre a data do ato e o classifique no exercício correto.
3. **Como Moderador (`MOD`) / Digitalizador de Acervo**, quero marcar a opção "Lançamento de Acervo Físico / Número Manual" e preencher o número que já consta no papel (ex: nº 14/2026), para que o arquivo digital tenha a mesma identificação do documento físico arquivado.
4. **Como Usuário Comum**, quero cadastrar novos atos diários no fluxo automático habitual, sem me preocupar com o próximo número, deixando que o sistema incremente atomicamente a partir do último número configurado ou emitido.

---

## Requirements

### Functional Requirements

#### 1. Modelo de Dados e Configuração de Sequências
- Criar a entidade/tabela `sequence_offsets`:
  - `id` (UUID, Chave Primária)
  - `municipality_id` (UUID, Foreign Key para `municipalities`, não nulo, indexado)
  - `type` (varchar(50), tipo do documento: `NOTICE`, `DECREE`, `ORDINANCE`, `LAW`, `CONTRACT`)
  - `contract_type` (varchar(50), nulo para tipos não contratuais; `service`, `bidding`, `publicinterest` para contratos)
  - `year` (integer, nulo para `LAW` que é perpétuo; obrigatório para os demais)
  - `initial_order` (integer, >= 1, número inicial a ser gerado quando não houver atos anteriores maiores)
  - `created_at`, `updated_at`
  - Restrição única composta: `UNIQUE(municipality_id, type, contract_type, year)`
- Criar endpoints da API de Sequências:
  - `GET /api/v1/municipalities/:id/sequences` — lista os marcos iniciais configurados do município (acessível a `MOD`).
  - `PUT /api/v1/municipalities/:id/sequences` — define ou atualiza o marco inicial para um par `(type, contract_type, year)`. Restrito exclusivamente a `MOD` (retorna 403 para `ADMIN` e `COMMON`).

#### 2. Criação de Documentos e Cálculo de Sequência (`order`)
- Atualizar `CreateDocumentInput` no backend:
  - `CreatedAt *time.Time` (opcional; se fornecido, valida data; se omitido, assume `time.Now()`).
  - `ManualOrder *int` (opcional, > 0; se fornecido, valida se o usuário autenticado possui role `MOD`. Se for `ADMIN` ou `COMMON`, retorna HTTP 403 Forbidden).
- Lógica de cálculo em transação atômica (`CreateWithNextOrder`):
  - Definir `year`: se `Type == LAW`, `year = nil`; caso contrário, `year = CreatedAt.Year()`.
  - Se `ManualOrder != nil`:
    - Verificar se já existe documento (`Unscoped`, incluindo soft-deleted) com `municipality_id`, `type`, `contract_type`, `year` e `order == *ManualOrder`.
    - Se existir, abortar transação e retornar erro `ErrOrderAlreadyExists` (HTTP 409 Conflict: "O número X já está cadastrado para este tipo e ano").
    - Se não existir, atribuir `d.Order = *ManualOrder`.
  - Se `ManualOrder == nil` (geração automática):
    - Obter `lastOrder = MAX(order)` existente no banco para aquele `municipality_id`, `type`, `contract_type` e `year`.
    - Obter `initialOrder` configurado em `sequence_offsets` para a combinação (padrão = 1 caso não configurado).
    - `d.Order = max(lastOrder, initialOrder - 1) + 1`.
  - Persistir o documento com `created_at = input.CreatedAt` (ou agora).

#### 3. Frontend — Formulário de Criação (`DocumentFormDialog.tsx`)
- Adicionar o seletor de **Data e Hora Oficial do Documento** no formulário de criação (já existente para edição, agora unificado na criação com default na data atual).
- Incluir chave/checkbox visual (apenas visível para usuários com role `MOD`):
  - `[ ] Lançamento de documento de acervo físico / Número manual`
  - Quando desmarcado (padrão):
    - Exibe banner/badge informativo sutil: *"O número sequencial será gerado automaticamente pelo sistema."*
  - Quando marcado:
    - Exibe campo numérico *"Número Oficial do Ato"*, com validação de número inteiro positivo maior que zero.
- Feedback de erro amigável na tela se a API retornar erro de duplicidade de número.

#### 4. Frontend — Tela de Configuração de Sequências
- Na área do município (visível e acessível exclusivamente para usuários com role `MOD`):
  - Aba ou seção "Sequências & Numeração".
  - Tabela listando os tipos de documentos, ano de exercício e campo para definir o marco inicial (`initial_order`).
  - Exibição em tempo real do *"Próximo número a ser gerado"*.

---

### Non-Functional Requirements

- **Segurança & RBAC:** Apenas perfis `MOD` (Moderadores) podem alterar a tabela de marcos iniciais (`sequence_offsets`) ou visualizar a tela de gestão de sequências. Apenas `MOD` pode informar `manualOrder` na criação. Usuários `ADMIN` e `COMMON` têm acesso bloqueado (HTTP 403 Forbidden).
- **Concorrência & Atomicidade:** Manter o lock transacional pessimista (`pg_advisory_xact_lock`) no cálculo do próximo número automático, impedindo race conditions em gravações simultâneas.
- **Performance:** As consultas de maior número e verificação de duplicidade devem utilizar os índices compostos existentes e o novo índice em `sequence_offsets`, respondendo em < 50ms.
- **Auditoria:** A data real em que o registro foi inserido no sistema pode ser rastreada (caso necessário, via logs estruturados `slog` com `request_id`, `creator_id` e dados do documento).
- **Testes:** 100% dos cenários (cálculo automático com offset, criação retroativa com ano passado, criação com número manual por MOD, bloqueio de ADMIN e COMMON para número manual/offset, colisão de número manual) cobertos por testes unitários e de integração.

---

## Acceptance Criteria

- [ ] Operador consegue selecionar a data oficial do ato no modal de criação de documentos.
- [ ] Documento cadastrado com data de ano anterior (ex: 2025) recebe numeração no ciclo do ano correspondente (2025) e não do ano corrente (2026).
- [ ] Moderador (`MOD`) consegue definir um marco inicial (offset) para um tipo de ato e ano (ex: Decretos 2026 iniciam em 85).
- [ ] Ao cadastrar o primeiro decreto com offset 85 configurado, o sistema atribui automaticamente `order = 85`.
- [ ] Leis municipais respeitam a sequência perpétua com offset configurado (sem reiniciar a cada ano).
- [ ] Usuários `MOD` conseguem marcar a opção de número manual e salvar um documento com número específico.
- [ ] Usuários `ADMIN` e `COMMON` não visualizam a opção de número manual nem a tela de configuração de sequências e recebem erro 403 caso tentem forçar via API.
- [ ] O sistema rejeita com HTTP 409 Conflict qualquer tentativa de gravar um número manual já existente no mesmo município/tipo/ano.
- [ ] Cadastros automáticos subsequentes após inserções manuais saltam para o próximo número livre (sem colidir com os manuais inseridos).
- [ ] Testes unitários no Go e Vitest no frontend cobrem todos os novos fluxos e passam sem falhas.

---

## Open Questions

- Nenhuma. O escopo da **Abordagem Completa 1** atende ponta a ponta tanto o início de operações no meio do ano quanto a digitalização de acervos legados e históricos.

---

## Success Metrics

- 0% de ocorrência de reinício acidental de contadores de atos para novos municípios implantados.
- 100% de precisão jurídica na associação entre a data oficial do ato e seu número sequencial anual.
- Redução a zero do retrabalho de suporte para correção manual de números de documentos via banco de dados.
