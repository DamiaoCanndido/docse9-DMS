---
name: startcycle
description: start cycle
---

# Start Cycle

**Type:** Workflow (orchestrates `@pm`, `@engeneer`, `@qa`, `@devops`)

## Purpose

Defines the standard end-to-end cycle for taking a request from idea to production, running each agent and its corresponding skill in sequence, with defined handoffs and feedback loops between phases.

## Invocation

```
/startcycle <idea>
```

- `<idea>` — a short description of the request, problem, or feature idea to run through the cycle (free text, one line or a short paragraph).
- Running this command starts the cycle at **Phase 1**, passing `<idea>` to `@pm` as the raw input for the `write_spec` skill.
- The cycle then proceeds automatically through Phases 2–4 unless a feedback loop sends it backward or the process is explicitly paused.

**Example:**

```
/startcycle Allow municipal staff to search empenhos by supplier name and date range
```

## Cycle Overview

```
@pm (write_spec)
   → @engeneer (generate_code)
   → @qa (audit_code)
   → @devops (deploy_app)
   → Released
```

---

## Phase 1 — Specify (`@pm` → `write_spec`)

**Trigger:** `/startcycle <idea>` is run.

**Runs:** `@pm` executes the `write_spec` skill.

**Output:** An approved spec (Problem, Goals, Non-Goals, Requirements, Acceptance Criteria).

**Exit condition:** Spec is marked "Approved." Open questions must be resolved — an unresolved question blocks handoff.

**Handoff to:** `@engeneer`

---

## Phase 2 — Build (`@engeneer` → `generate_code`)

**Trigger:** Approved spec from Phase 1.

**Runs:** `@engeneer` executes the `generate_code` skill.

**Output:** Implemented code (backend + frontend as needed) with accompanying unit tests, following stack conventions.

**Exit condition:** Code compiles, unit tests pass locally, and self-review is complete.

**Handoff to:** `@qa`

**Feedback loop:** If the spec turns out ambiguous or contradictory during implementation, return to `@pm` before proceeding — do not silently interpret.

---

## Phase 3 — Verify (`@qa` → `audit_code`)

**Trigger:** Implemented code from Phase 2.

**Runs:** `@qa` executes the `audit_code` skill.

**Output:** Either a sign-off (ready to ship) or one or more filed bug reports.

**Exit condition:** All blocking issues resolved; acceptance criteria from the spec are verified as met.

**Handoff to:** `@devops`

**Feedback loop:** Blocking bugs send the cycle back to Phase 2 (`@engeneer`). If the issue is actually a spec gap (behavior undefined), send it back to Phase 1 (`@pm`) instead of guessing.

---

## Phase 4 — Ship (`@devops` → `deploy_app`)

**Trigger:** `@qa` sign-off from Phase 3.

**Runs:** `@devops` executes the `deploy_app` skill.

**Output:** Deployed change in production, verified via health checks/smoke tests, with monitoring active.

**Exit condition:** Deployment verified stable; rollback plan was ready throughout.

**Feedback loop:** A failed deployment triggers an immediate rollback, not a forward fix — the cycle re-enters at Phase 2 with the incident as new context.

---

## Cycle Tracker

```markdown
## Cycle: [name/ticket]

- [ ] Phase 1 — Spec approved by @pm (link)
- [ ] Phase 2 — Code implemented by @engeneer (link/PR)
- [ ] Phase 3 — Verified by @qa (sign-off or bugs filed)
- [ ] Phase 4 — Deployed by @devops (link/release)

**Current phase:**
**Blocked on:**
**Notes:**
```

## Guidelines

- Each phase has exactly one owner agent at a time — no phase starts before the previous one's exit condition is met.
- Feedback loops go backward only as far as necessary — a code bug doesn't need to restart the spec, but a spec gap does need to restart there.
- Skipping a phase (e.g., deploying without `@qa` sign-off) is only acceptable for pre-agreed emergency fixes, and must be documented as such in the Cycle Tracker.
- The cycle can run for a small fix in minutes or a large feature over days — the sequence doesn't change, only the depth of each phase.