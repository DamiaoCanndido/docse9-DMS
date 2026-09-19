# Merge da Branch develop na Branch master — Spec

**Status:** Approved  
**Author:** @pm  
**Date:** 2026-09-19  
**Related:** `.agents/workflows/startcycle.md`, `docs/specs/whatsnew-contract-monitoring.md`

## Problem Statement
O ciclo anterior implementou o ajuste informativo de monitoramento de contratos (1 semana antes) no componente `WhatsNewAlertDialog`, com seus testes e documentação na branch `develop`. É necessário consolidar essas alterações na branch estável de produção `master` através de um processo formal e seguro de merge.

## Goals
- Comitar as alterações aprovadas na branch `develop`.
- Realizar o merge seguro da branch `develop` na branch `master`.
- Validar a integridade da aplicação após o merge executando testes unitários e build de produção.

## Non-Goals
- Não adicionar novas features ou alterações não testadas diretamente na `master`.
- Não alterar configurações de deploy em produção fora do escopo do merge.

## User Stories
- Como gestor e operador do Docseq, quero que a branch `master` contenha a versão mais recente e estável do código presente em `develop`, garantindo paridade entre desenvolvimento e produção.

## Requirements
### Functional
1. As modificações em `WhatsNewAlertDialog.tsx`, `WhatsNewAlertDialog.test.tsx` e `docs/specs/whatsnew-contract-monitoring.md` devem ser comitadas em `develop`.
2. A branch `master` deve receber o merge da branch `develop` gerando um merge commit padronizado.
3. Não deve haver conflitos de merge.

### Non-Functional
- Zero erros na suíte de testes do frontend (Vitest).
- Zero erros na suíte de testes do backend (Go).
- Build de produção (`npm run build`) sem falhas.

## Acceptance Criteria
- [x] Árvore de trabalho limpa e alterações comitadas em `develop`.
- [x] Merge de `develop` para `master` realizado com sucesso.
- [x] Todos os testes no frontend e backend executados e aprovados na `master`.
- [x] Build de produção compilado com sucesso na `master`.
