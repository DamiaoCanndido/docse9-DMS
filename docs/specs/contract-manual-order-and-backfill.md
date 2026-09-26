# Gestão de Acervo Físico, Número Manual e Data de Vigência em Contratos — Spec

**Status:** Approved  
**Author:** @pm  
**Date:** 2026-09-26  
**Related:** `docs/specs/flexible-document-sequences-and-backfill.md`, `frontend/src/components/documents/DocumentFormDialog.tsx`, `frontend/src/components/__tests__/DocumentFormDialog.test.tsx`, `backend/internal/repository/document_repository.go`, `backend/internal/service/document_service.go`

---

## Problem Statement

Atualmente, o diálogo de criação de documentos (`DocumentFormDialog.tsx`) restringe a seleção de número manual para contratos através da trava `type !== 'CONTRACT'`, impedindo que Moderadores (`MOD`) cadastrem contratos legados e digitalizem acervos físicos preexistentes. Além disso, contratos novos não enviam a data de registro oficial (`createdAt`) no payload, fazendo com que o backend atribua o relógio atual do servidor e vincule o exercício da numeração anual ao ano corrente em vez do ano de vigência do contrato. Por fim, a inclusão do bloco de acervo físico no formulário de contratos exige um balanceamento do layout em duas colunas para preservar a usabilidade sem estourar o viewport vertical em telas de menor densidade.

---

## Goals

1. **Habilitar Lançamento de Acervo Físico / Número Manual para Contratos:** Permitir exclusivamente a Moderadores (`MOD`) informar manualmente o número (`order`) de contratos físicos legados durante a digitalização de acervos.
2. **Definição da Data Oficial via Data de Início (`startIn`):** Utilizar a Data de Início da vigência (`startIn`) como `createdAt` do contrato no envio do formulário de criação (Opção A), assegurando que a sequência anual (`year`) seja particionada no exercício financeiro correto do contrato.
3. **Harmonização do Layout Responsivo:** Ajustar o grid de duas colunas do `DocumentFormDialog` para acomodar os campos de upload, acervo físico e detalhes do contrato de forma compacta e equilibrada, sem gerar scroll vertical excessivo em resoluções padrão (>= 768px de altura).
4. **Validação de Unicidade e Conflitos:** Manter a consistência transacional do backend, exibindo mensagens claras caso um número manual já exista no município para o mesmo tipo de contrato e exercício.

---

## Non-Goals

- Não alterar a arquitetura nem o schema do backend: os endpoints, a tabela `sequence_offsets` e o repositório Go já suportam `type = CONTRACT`, subtipos de contrato e `manualOrder`.
- Não criar campos adicionais redundantes de data: a data oficial do contrato derivará de forma direta da "Data de Início" (`startIn`) já preenchida pelo operador (Opção A).
- Não permitir que usuários com perfil `ADMIN` ou `COMMON` insiram número manual em contratos (privilégio estritamente restrito a `MOD`).
- Não permitir alteração do número (`order`) de contratos após a sua criação (imutabilidade jurídica após o registro oficial).

---

## Background / Context

Na administração pública municipal (obedecendo às Leis Federais nº 4.320/64 e nº 14.133/21), os contratos administrativos possuem numeração sequencial anual renovada a cada 1º de janeiro e particionada por modalidade/categoria (Prestação de Serviços, Licitação e Interesse Público). 

Na especificação original de sequências flexíveis (`docs/specs/flexible-document-sequences-and-backfill.md`), todo o backend foi desenhado e implementado para suportar `manualOrder` e marcos iniciais (`initial_order`) em contratos. No entanto, no frontend, por herança do layout prévio que tratava contratos de forma segregada dos demais atos, o checkbox de acervo físico recebeu uma condição de exclusão (`type !== 'CONTRACT'`). Com isso, equipes municipais em fase de implantação ficaram impossibilitadas de digitalizar contratos físicos anteriores e importar séries históricas.

---

## User Stories

1. **Como Moderador (`MOD`)**, quero marcar a opção "Lançamento de documento de acervo físico / Número manual" ao cadastrar um contrato físico preexistente (ex: Contrato nº 14/2024 de Prestação de Serviços) para que o sistema preserve fielmente o identificador original do documento.
2. **Como Operador / Moderador**, quero cadastrar um contrato informando sua Data de Início (ex: 10/02/2024), garantindo que o sistema registre esse ato no exercício de 2024 e incremente a sequência anual de 2024, e não a do ano corrente (2026).
3. **Como Usuário Comum**, quero cadastrar um contrato e visualizar a indicação clara de que o número será gerado automaticamente pelo sistema, sem acesso à alteração manual.
4. **Como Usuário do Sistema**, quero interagir com o modal de contratos em computadores com telas de menor resolução (ex: 1366x768) de forma fluida e sem barras de rolagem desnecessárias.

---

## Requirements

### Functional Requirements

#### 1. Remoção da Restrição de Número Manual no Formulário
- No componente [`DocumentFormDialog.tsx`](file:///home/nergal/apps/docSe9-DMS/frontend/src/components/documents/DocumentFormDialog.tsx), a condicional de renderização do bloco de acervo físico deve ser alterada de:
  ```tsx
  {!editingDocument && isMod && type !== 'CONTRACT' && (
  ```
  para:
  ```tsx
  {!editingDocument && isMod && (
  ```
- O checkbox *"Lançamento de documento de acervo físico / Número manual"* e seu respectivo input numérico *"Número Oficial do Ato"* devem estar disponíveis para todos os tipos documentais, incluindo `CONTRACT`.
- Quando desmarcado, exibe o aviso: *"O número sequencial será gerado automaticamente pelo sistema."*
- Quando marcado, o campo numérico torna-se obrigatório e deve aceitar apenas inteiros positivos maiores que zero (`min={1}`).

#### 2. Sincronização da Data Oficial do Contrato (Opção A)
- Na submissão de um novo contrato (`type === 'CONTRACT'` no fluxo de criação):
  - Combinar `startInDate` e `startInTime` gerando a data e hora ISO completa através de `combineDateAndTime(startInDate, startInTime)`.
  - Atribuir esse valor a `createInput.createdAt`, além de `createInput.startIn`:
    ```ts
    if (type === 'CONTRACT') {
      const contractDateTime = combineDateAndTime(startInDate, startInTime);
      createInput.contractType = contractType;
      createInput.value = Number(value);
      createInput.duration = Number(duration);
      createInput.startIn = contractDateTime;
      createInput.createdAt = contractDateTime;
    }
    ```
  - Dessa forma, o backend extrairá o ano do contrato a partir de `docCreatedAt.Year()`, associando corretamente a numeração ao exercício da vigência contratual.

#### 3. Envio de Número Manual para Contratos
- Se `isManualOrder` estiver marcado e preenchido, preencher `createInput.manualOrder = Number(manualOrder)` tanto para documentos convencionais quanto para contratos.
- Em caso de duplicidade de número (`HTTP 409 Conflict`), exibir a mensagem de erro amigável na base do formulário, respeitando o padrão existente: *"O número informado já está cadastrado para este tipo e ano."*

#### 4. Harmonização de Layout do Diálogo (Grid de Duas Colunas)
- Na Coluna 1 (Identificação & Arquivo):
  - Tipo de Documento (`Select`).
  - Dropzone de Arquivo / Upload R2.
  - Para `type !== 'CONTRACT'`: Bloco de Data e Hora de Registro Oficial.
  - Para todos os tipos (quando `!editingDocument && isMod`): Bloco de Lançamento de Acervo Físico / Número Manual.
- Na Coluna 2 (Conteúdo & Detalhes):
  - Campo Descrição / Ementa:
    - Quando `type === 'CONTRACT'`, manter a altura compacta da `textarea` (ex: `h-16 md:h-20`).
    - Quando `type !== 'CONTRACT'`, manter a altura expandida (`h-24 md:h-[156px]`).
  - Bloco de Detalhes do Contrato (`type === 'CONTRACT'`):
    - Reduzir ligeiramente `padding` e espaçamentos internos (`p-2.5`, `gap-2`) para assegurar que a altura da Coluna 2 equilibre-se perfeitamente com a Coluna 1.
- Manter o limite de altura total do diálogo contido na viewport padrão (768px de altura), sem barra de rolagem vertical involuntária no corpo do modal.

---

### Non-Functional Requirements

- **Segurança & RBAC:** Apenas usuários autenticados com papel `MOD` podem visualizar o campo de número manual e enviar `manualOrder`. Usuários `ADMIN` e `COMMON` não visualizam a opção e têm qualquer tentativa bloqueada pela API com HTTP 403.
- **Concorrência & Atomicidade:** Operações de criação de contratos manuais ou automáticos devem continuar utilizando os locks pessimistas (`pg_advisory_xact_lock`) e índices compostos do PostgreSQL já ativos.
- **Compatibilidade:** O layout deve responder responsivamente: em telas pequenas (< 768px), organizar em coluna única vertical; em telas médias e grandes (>= 768px), organizar em duas colunas simétricas.
- **Qualidade & Testes:** Os testes unitários do frontend (`DocumentFormDialog.test.tsx`) devem cobrir especificamente a presença e o comportamento do número manual em contratos, bem como a sincronização do `createdAt`.

---

## Out of Scope

- Edição de numeração (`order`) para contratos já cadastrados.
- Inclusão de um terceiro campo de data (ex: data de assinatura separada de data de vigência).
- Alteração nos filtros da listagem de contratos ou nos endpoints de sequências da API.

---

## Acceptance Criteria

- [x] Usuário Moderador (`MOD`) visualiza a opção "Lançamento de documento de acervo físico / Número manual" ao selecionar o tipo "Contrato" no modal de criação.
- [x] Usuários `ADMIN` e `COMMON` não visualizam a opção de número manual ao selecionar o tipo "Contrato".
- [x] Ao marcar "Lançamento de acervo físico" para um contrato e preencher o número (ex: 45), o payload enviado ao backend contém `manualOrder: 45`.
- [x] Ao cadastrar um contrato com Data de Início em ano anterior (ex: 15/05/2024), o payload enviado contém `createdAt` correspondente a essa data, e o backend calcula/valida a sequência no exercício de 2024.
- [x] O sistema rejeita e exibe erro amigável caso o Moderador tente cadastrar um contrato manual com número já existente para aquele tipo de contrato e ano.
- [x] O diálogo de criação de contrato acomoda todos os campos sem quebras visuais e sem gerar rolagem de página vertical em resolução de 1366x768.
- [x] A suíte de testes do frontend passa com 100% de sucesso, incluindo novos testes unitários para número manual e payload em contratos.

---

## Success Metrics

- 100% de paridade funcional entre Contratos e demais atos oficiais no suporte a acervo físico legado.
- Zero ocorrências de contratos históricos gravados no exercício financeiro incorreto devido ao relógio do servidor.
- Layout responsivo estável e sem overflow vertical indesejado em telas de notebooks.
