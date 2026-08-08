---
name: deploy-app
description: >-
  Defines how the DevOps Master agent (@devops) takes approved code and ships it to production safely, reliably, and with a clear rollback path. Use this skill when deploying a release, setting up environments, rotating secrets, or responding to failed deployments.
---

# Deploy App

**Skill for:** `@devops`

## Purpose

Defines how the DevOps Master agent takes code approved by `@qa` and ships it to production safely, reliably, and with a clear rollback path.

## When to use

- A release is ready to go out after passing `@qa`'s audit
- Setting up a new environment (dev, staging, prod)
- Rotating secrets or updating infrastructure configuration
- Responding to a failed deployment or production incident

## Process

1. **Confirm readiness** — Verify `@qa` has signed off and the target commit/branch is the intended one.
2. **Review the diff for infra impact** — New environment variables, migrations, dependencies, or config changes that need to ship alongside the code.
3. **Build and tag** — Build the Docker image, tag it with the commit SHA (never `latest` alone), and push to the registry.
4. **Run migrations first** — Apply database migrations before switching traffic to the new version; ensure they're backward-compatible with the version being replaced.
5. **Deploy** — Roll out using the environment's strategy (rolling update / blue-green); avoid downtime where the setup allows it.
6. **Verify** — Run health checks and smoke tests against the new deployment before considering it complete.
7. **Monitor** — Watch logs and error rates for a defined window post-deploy; know the rollback command before you need it.
8. **Document** — Update the runbook with anything that deviated from the standard process.

## Infrastructure Conventions

- **Containers:** Docker Compose mirrors prod topology in dev; no environment-specific code paths — only environment-specific config.
- **Servers:** Debian/Ubuntu, SSH key-only auth, Fail2ban enabled, firewall default-deny with explicit allow rules.
- **File transfer:** SFTP via chrooted `internal-sftp`, dedicated non-shell users per service.
- **Secrets:** Never committed to the repo; injected via environment variables or a secrets manager, rotated on a schedule and immediately on suspected exposure.
- **Backups:** Automated, tested restores (not just backup jobs that "should" work) — rsync/Borgbackup with off-site copies.
- **Database:** Connections over TLS where possible, least-privilege DB users per service, migrations are additive/backward-compatible by default.

## Rollback Checklist

```markdown
- [ ] Previous image tag/version identified
- [ ] Rollback command tested and ready
- [ ] Migration rollback plan exists (or migration was backward-compatible)
- [ ] Stakeholders notified of the rollback
- [ ] Root cause logged for post-mortem
```

## Guidelines

- Never deploy on a Friday afternoon (or right before you're unreachable) unless it's a critical fix.
- If a migration can't be made backward-compatible, deploy it as two releases: additive change first, cleanup after.
- Treat monitoring/alerting as part of the deploy, not an afterthought — a silent failure is worse than a loud one.
- Prefer the boring, well-understood deployment method over a novel one, especially under time pressure.
- Every incident gets a runbook update; the goal is that it never repeats the same way twice.
