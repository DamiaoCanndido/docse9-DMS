---
name: generate-code
description: >-
  Defines how the Full Stack Engineer agent (@engeneer) turns a spec into working, tested, maintainable code. Use this skill when implementing a new feature, fixing a bug, refactoring code, or building services, endpoints, or components.
---

# Generate Code

**Skill for:** `@engeneer`

## Purpose

Defines how the Full Stack Engineer agent turns a spec (from `@pm`) into working, tested, maintainable code — covering both backend and frontend.

## When to use

- Implementing a new feature from an approved spec
- Fixing a bug
- Refactoring existing code
- Building a new service, endpoint, or component

## Process

1. **Read the spec** — Confirm requirements, acceptance criteria, and non-goals from `@pm` before writing code. Flag ambiguity instead of guessing.
2. **Design first** — Sketch the data model, API contract (request/response shapes), and where the change fits in the existing architecture.
3. **Implement backend** — Follow the layered architecture: `handler` (HTTP, validation) → `service` (business logic) → `repository` (persistence). Keep layers isolated and testable.
4. **Implement frontend** — Build components against the API contract; handle loading, error, and empty states, not just the happy path.
5. **Write tests alongside the code** — Don't defer testing to `@qa` entirely; ship code with unit coverage for the logic you just wrote.
6. **Self-review** — Re-read the diff as if reviewing someone else's PR. Check for dead code, missing error handling, and unclear naming.
7. **Document** — Update the OpenAPI spec for new/changed endpoints; add code comments only where intent isn't obvious from the code itself.
8. **Hand off** — Summarize what changed, why, and any trade-offs made, for `@qa` and `@devops`.

## Stack Conventions

**Backend (Go + Gin + PostgreSQL):**
- Clean Architecture: `handler → service → repository`, dependencies point inward
- GORM for persistence; `pgconn.PgError` for constraint/unique-violation handling
- JWT auth via middleware (e.g. `RequireRole`); never trust client-side role claims alone
- UUID v4 for primary keys
- Errors are wrapped with context (`fmt.Errorf("...: %w", err)`), not swallowed
- New/changed endpoints get an OpenAPI 3.0 entry

**Frontend (Next.js + React + TypeScript + tailwindcss v4 + ZOD):**
- Validate forms with Zod + React Hook Form; never rely on client-only validation for anything security- or data-integrity-relevant
- Debounce user input before triggering search/filter requests
- Sync filter/search state to URL params so views are shareable/bookmarkable
- Prefer typed API clients over ad-hoc `fetch` calls scattered across components, use axios lib.
- Prioritize responsive design (mobile-first) and accessibility (WCAG 2.1 AA)

## Guidelines

- Match existing patterns in the codebase before introducing a new one; consistency beats personal preference.
- No feature is "done" without tests — untested code is a liability, not a deliverable.
- Prefer small, reviewable PRs over large ones; if a change can't be described in one sentence, consider splitting it.
- Never commit secrets, credentials, or `.env` files.
- If a requirement is unclear or contradicts existing behavior, ask `@pm` before implementing an assumption.
