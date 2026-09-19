# Ajuste no WhatsNewAlertDialog — Monitoramento de Contratos — Spec

**Status:** Approved  
**Author:** @pm  
**Date:** 2026-09-19  
**Related:** `.agents/workflows/startcycle.md`, `frontend/src/components/WhatsNewAlertDialog.tsx`, `frontend/src/components/NotificationCenter.tsx`

## Problem Statement
O componente `WhatsNewAlertDialog` atualmente exibe no card de Alertas de Vigência de Contratos o texto legadov dizendo que os contratos são alertados com "30, 60 e 90 dias de antecedência". No entanto, a implementação atual da plataforma Docseq monitora e alerta contratos com **1 semana de antecedência (7 dias)**. Essa divergência gera confusão para os gestores municipais sobre quando receberão os alertas no sistema.

## Goals
- Atualizar a descrição do recurso "Alertas de Vigência de Contratos" no componente `WhatsNewAlertDialog` para informar que o monitoramento e alerta ocorrem uma semana antes (7 dias de antecedência).
- Atualizar a suíte de testes unitários do componente (`WhatsNewAlertDialog.test.tsx`) para garantir a asserção do novo texto.

## Non-Goals
- Não alterar a regra de negócio de cálculo de vencimento em `NotificationCenter.tsx` ou no backend.
- Não alterar outros cards ou textos não relacionados no modal de novidades.

## Background / Context
Na versão v1.1.0 (commit `3a39391`), a Central de Notificações consolidou a janela de alerta para contratos a vencer em até 7 dias (uma semana) e contratos expirados recentemente. O componente `WhatsNewAlertDialog` precisava ser sincronizado com essa regra.

## User Stories
- Como gestor municipal (`MOD` / `USER`), quero ser informado no modal de novidades que o sistema monitora os contratos com uma semana de antecedência, para que eu possa me planejar adequadamente para aditivos e renovações.

## Requirements

### Functional
1. No componente `WhatsNewAlertDialog`, o parágrafo descritivo de "Alertas de Vigência de Contratos" deve informar que o sistema alerta com uma semana de antecedência (7 dias) para renovações e aditivos.
2. Manter a estilização, ícones e comportamento de dismiss idênticos aos existentes.

### Non-Functional
- O texto deve manter a harmonia visual, sem quebras de layout na caixa de diálogo.
- Sem dependências adicionais ou overhead de renderização.

## Acceptance Criteria
- [x] O texto de `WhatsNewAlertDialog` menciona claramente o monitoramento com uma semana de antecedência.
- [x] Os testes unitários em `WhatsNewAlertDialog.test.tsx` cobrem o novo texto e passam com sucesso.
- [x] O build de produção do frontend (`npm run build`) compila sem erros.
- [x] As funcionalidades existentes de exibição única por versão e persistência no `localStorage` permanecem intactas.

## Open Questions
- Nenhuma. O requisito é objetivo e alinhado ao comportamento do sistema.

## Success Metrics
- 100% de precisão na comunicação de novidades para os usuários finais sobre o ciclo de alertas de contratos.
