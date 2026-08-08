---
name: audit-code
description: >-
  Defines how the QA Engineer agent (@qa) verifies that implemented code meets the spec and is safe to ship. Use this skill when reviewing code, writing test plans, or investigating bugs.
---

# Audit Code

**Skill for:** `@qa`

## Purpose

Defines how the QA Engineer agent verifies that implemented code meets the spec (from `@pm`) and is safe to ship — through test design, execution, and exploratory review.

## When to use

- Code from `@engeneer` is ready for review before merge/release
- A bug has been reported and needs investigation
- A feature needs a test plan before implementation starts
- Verifying a fix doesn't reintroduce a regression

## Process

1. **Re-read the spec** — Pull acceptance criteria from `@pm`'s spec; these are the pass/fail bar, not a suggestion.
2. **Map coverage to the test pyramid** — Confirm unit, integration, and (where relevant) end-to-end tests exist for the change; flag gaps instead of filling them silently.
3. **Run the test suite** — Confirm everything passes locally/in CI before further manual review.
4. **Exploratory testing** — Go beyond the acceptance criteria: invalid input, empty/null states, concurrent access, boundary values, permission edge cases.
5. **Contract check** — Verify API responses match the OpenAPI spec (status codes, shapes, error formats).
6. **Reproduce and isolate bugs** — Before filing, confirm the bug is reproducible and narrow it to the smallest reliable repro.
7. **File or clear** — Report blocking issues with full repro steps, or confirm the change is ready to ship.

## Test Coverage Checklist

**Unit (testify/mock):**
- Business logic in the `service` layer, isolated from HTTP and DB
- Edge cases: nil/empty inputs, boundary values, error paths

**Handler (httptest):**
- Request validation and status codes
- Auth/role enforcement (`RequireRole` middleware) rejects unauthorized requests
- Error responses match the documented shape

**Integration (testcontainers-go):**
- Real PostgreSQL interactions: constraints, unique violations, transactions
- Multi-step flows that cross layers (handler → service → repository)

## Bug Report Template

```markdown
**Title:** [short, specific summary]

**Severity:** Blocker | Major | Minor | Cosmetic

**Steps to Reproduce:**
1. ...
2. ...

**Expected:**
...

**Actual:**
...

**Environment:** (branch/commit, env, data state)

**Notes:** (screenshots, logs, related spec/PR)
```

## Guidelines

- A missing test on security- or data-integrity-relevant code is a blocker, not just a code smell.
- Separate "doesn't match the spec" from "I'd have designed it differently" — only the former blocks a release.
- Always attach a minimal reproduction; "it doesn't work" isn't an actionable bug report.
- Test the failure paths as thoroughly as the happy path — most production incidents live there.
- When in doubt about expected behavior, check with `@pm` before deciding it's a bug.
