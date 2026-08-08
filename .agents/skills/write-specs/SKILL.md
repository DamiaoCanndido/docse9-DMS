---
name: write-specs
description: >-
  Defines the process and template the Product Manager agent (@pm) uses to turn a raw idea, request, or problem into a clear, actionable specification. Use this skill when scoping features, refining stakeholder requests, or documenting a redesign.
---

# Write Specs

**Skill for:** `@pm`

## Purpose

Defines the process and template the Product Manager agent uses to turn a raw idea, request, or problem into a clear, actionable specification before engineering work begins.

## When to use

- A new feature or product is being considered
- A stakeholder request needs to be turned into requirements
- Scope needs to be locked down before `@engeneer` starts implementation
- An existing feature needs a documented redesign

## Process

1. **Clarify the problem** — What's broken or missing? Who is affected, and how often?
2. **Define the goal** — What does success look like? What should NOT be attempted (non-goals)?
3. **Identify users and use cases** — Who uses this, and in what scenarios?
4. **Write requirements** — Separate functional (what it does) from non-functional (performance, security, compliance).
5. **Set acceptance criteria** — Concrete, testable conditions for "done." These feed directly into `@qa`'s test plan.
6. **Flag open questions and risks** — Anything unresolved, ambiguous, or dependent on someone else.
7. **Define success metrics** — How will we know this worked after shipping?
8. **Review before handoff** — Confirm scope with stakeholders before passing to `@engeneer`.

## Spec Template

```markdown
# [Feature/Product Name] — Spec

**Status:** Draft | In Review | Approved
**Author:**
**Date:**
**Related:** (links to tickets, discussions, prior specs)

## Problem Statement
What problem are we solving, and for whom? Why does it matter now?

## Goals
- ...

## Non-Goals
- ...

## Background / Context
Relevant history, constraints, prior attempts, data supporting the need.

## User Stories
- As a [user], I want [goal], so that [benefit].

## Requirements

### Functional
1. ...

### Non-Functional
- Performance: ...
- Security / Compliance: ...
- Availability: ...

## Out of Scope
- ...

## Acceptance Criteria
- [ ] ...
- [ ] ...

## Open Questions
- ...

## Success Metrics
How we'll measure whether this achieved its goal post-launch.

## Rollout Plan (optional)
Phased release, feature flags, migration steps, etc.
```

## Guidelines

- Keep the problem statement to 2–3 sentences — if it needs more, the problem isn't clear yet.
- Every requirement should be testable; if `@qa` can't write a test for it, rewrite it.
- Non-goals are as important as goals — they prevent scope creep during implementation.
- Don't specify implementation details (that's `@engeneer`'s call) — specify outcomes and constraints, not solutions.
- A spec is done when someone outside the conversation could read it and know exactly what's being built and why.
