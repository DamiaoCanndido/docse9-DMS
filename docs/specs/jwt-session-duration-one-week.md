# Token JWT com Duração de 1 Semana e Notificação no WhatsNewAlertDialog — Spec

**Status:** Approved  
**Author:** @pm  
**Date:** 2026-09-19  
**Related:** `.agents/workflows/startcycle.md`, `backend/internal/service/auth_service.go`, `backend/internal/handler/user_handler.go`, `frontend/src/app/api/auth.ts`, `frontend/src/components/WhatsNewAlertDialog.tsx`

## Problem Statement
Atualmente, os tokens JWT emitidos no login e na alteração de senha expiram em 24 horas (1 dia). Essa janela exige que os operadores e gestores municipais façam login diariamente, interrompendo o fluxo contínuo de trabalho administrativo na gestão de documentos e contratos. Além disso, quando essa melhoria for disponibilizada, os usuários precisam ser avisados transparentemente através do modal de novidades (`WhatsNewAlertDialog`).

## Goals
- Aumentar o tempo de expiração do token JWT de 24 horas para **1 semana (7 dias / 168 horas)** no backend.
- Sincronizar o tempo de vida do cookie de sessão (`maxAge`) no frontend para **7 dias (604.800 segundos)**.
- Notificar os usuários sobre a nova duração de sessão de 7 dias dentro do componente `WhatsNewAlertDialog`.
- Incrementar a versão da aplicação (para `v1.2.0`) garantindo que todos os usuários recebam o alerta de novidades atualizado mesmo que já tenham dispensado a versão anterior.
- Atualizar os testes unitários e de integração no backend e frontend para validar as alterações.

## Non-Goals
- Não implementar sistema de refresh tokens complexo ou persistência de sessões em Redis/banco neste ciclo (o token JWT autocontido assinado com HMAC-SHA256 atende plenamente o escopo).
- Não alterar as claims de usuário (`user_id`, `username`, `role`, `municipality_id`, `must_change_password`).

## Background / Context
Os servidores municipais utilizam o Docseq em dias úteis consecutivos e relataram atrito com a necessidade de reautenticação diária. Uma duração de 7 dias equilibra comodidade operacional e segurança (o token continua protegido por `httpOnly`, `secure` em produção e `sameSite: 'lax'`).

## User Stories
- Como operador ou gestor municipal, quero que minha sessão permaneça válida por até 1 semana para não precisar redigitar minhas credenciais a cada dia de trabalho.
- Como usuário do sistema, quero ser notificado no diálogo de novidades ("O que há de novo") sobre a nova duração da sessão para saber que meu login permanecerá ativo por mais tempo.

## Requirements

### Functional
1. **Backend**:
   - Em `backend/internal/service/auth_service.go`, gerar o token com duração de `7 * 24 * time.Hour` (168h) no método de login (`Login`).
   - Em `backend/internal/handler/user_handler.go`, gerar o token com duração de `7 * 24 * time.Hour` no método de alteração obrigatória de senha (`ForceChangePassword`).
2. **Frontend**:
   - Em `frontend/src/app/api/auth.ts`, definir `maxAge: 7 * 24 * 60 * 60` (604.800s) para os cookies de autenticação (`token` e `user`) no login, no reset de senha e na atualização de perfil.
   - Em `frontend/src/components/WhatsNewAlertDialog.tsx`, incluir destaque visual informando a "Sessão Estendida de 7 Dias".
   - Em `frontend/src/components/NotificationCenter.tsx`, adicionar a novidade no feed `SYSTEM_CHANGELOG_NOTIFICATIONS`.
   - Atualizar a versão para `v1.2.0` em `frontend/package.json` e `frontend/src/lib/version.ts`.

### Non-Functional
- Manter padrões de segurança para cookies: `httpOnly: true`, `secure: true` em produção, `sameSite: 'lax'`.
- Garantir que `WhatsNewAlertDialog` mantenha layout responsivo com rolagem suave caso o número de novidades ultrapasse a viewport do modal.
- 100% de aprovação na suíte de testes do backend e frontend.

## Acceptance Criteria
- [ ] Backend gera tokens JWT com claim `exp` correspondente a 7 dias a partir da emissão.
- [ ] Cookies no frontend configurados com validade de 7 dias.
- [ ] `WhatsNewAlertDialog` exibe o card informando sobre a validade de 7 dias da sessão.
- [ ] A chave de visualização no `localStorage` é atualizada com a nova versão, garantindo exibição para os usuários.
- [ ] Testes unitários do frontend (`WhatsNewAlertDialog.test.tsx`) cobrem a nova notificação.
- [ ] Testes do backend e frontend compilam e passam com sucesso.

## Open Questions
- Nenhuma. Escopo bem delineado e aprovado.

## Success Metrics
- Redução substancial de reautenticações diárias repetitivas por parte dos operadores.
- Notificação efetiva exibida a 100% dos usuários ativos após a implantação.
