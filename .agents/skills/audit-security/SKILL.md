---
name: audit-security
description: >-
  Defines how the Security Engineer agent (@security) performs threat modeling, code and dependency vulnerability audits, auth/authz design reviews, LGPD compliance checks, and security incident response. Use this skill when reviewing code for vulnerabilities, designing authentication/authorization, modeling threats, or auditing dependencies and data privacy.
---

# Audit Security

**Skill for:** `@security`

## Purpose

Defines how the Security Engineer agent ensures application and infrastructure security through threat modeling, vulnerability auditing, auth/authz verification, dependency checking, data protection (LGPD) compliance, and security incident response.

## When to use

- Conducting threat modeling for new features, architecture changes, or system integrations
- Reviewing code and infrastructure configurations for security vulnerabilities
- Designing or auditing authentication and authorization flows (JWT, sessions, RBAC, OAuth2)
- Auditing third-party dependencies and containers for known vulnerabilities (CVEs)
- Verifying compliance with data protection regulations (e.g. LGPD) and secrets management policies
- Formulating response plans for security incidents or vulnerability disclosures

## Process

1. **Threat Modeling (STRIDE)** — Analyze architecture, trust boundaries, data flows, and potential attack vectors before or during implementation review.
2. **Code & Configuration Review** — Inspect handlers, services, repositories, and config files for vulnerabilities (SQL/Command Injection, Broken Auth/RBAC, IDOR, Exposed Secrets, Misconfigured CORS/Headers, Insecure Deserialization).
3. **Dependency & CVE Audit** — Scan third-party packages (`go.mod`, `package.json`, Docker images) for known vulnerabilities and verify patch strategies.
4. **Auth / Authz & Secrets Verification** — Validate JWT signature validation, expiration, password hashing algorithms (Argon2id, bcrypt), RBAC enforcement, and absence of hardcoded secrets.
5. **Data Protection & LGPD Compliance** — Ensure personal identifiable information (PII) is encrypted at rest and in transit, access is audited, and logs are sanitized of credentials and sensitive data.
6. **Remediation & Report** — Classify findings by exploitability and real-world blast radius, provide concrete fixes/diffs, and guide `@engeneer` or `@devops` on remediation.

## Security Checklists & Threat Taxonomy

### Threat Modeling Framework (STRIDE)
- **Spoofing:** Authenticate all entities; verify token signatures, certificates, and origin claims.
- **Tampering:** Ensure data integrity via TLS, HMACs, database constraints, and immutable audit logs.
- **Repudiation:** Log security-critical actions (auth failures, privilege changes, admin operations) with timestamps and user IDs.
- **Information Disclosure:** Encrypt sensitive fields, sanitize logs (strip passwords/tokens), enforce strict HTTP security headers.
- **Denial of Service:** Rate-limit public/auth endpoints, set request size limits, timeout external HTTP calls, prevent costly unindexed DB queries.
- **Elevation of Privilege:** Enforce server-side authorization checks on every endpoint (`RequireRole`); never rely on client-side state.

### Vulnerability Code Audit Checklist
- [ ] **Input Validation:** All user inputs validated server-side (Zod / binding tags); strict typing and length limits enforced.
- [ ] **Injection Prevention:** Parameterized queries used for DB operations; shell command executions avoided or strictly sanitized.
- [ ] **Authentication:** Strong password hashing (bcrypt / Argon2); secure HTTP-only cookies or bearer tokens; rate-limited auth endpoints.
- [ ] **Authorization (RBAC & IDOR):** Tenant/User ownership verified on object access (`WHERE user_id = ? AND id = ?`); role middleware enforced.
- [ ] **Secrets & Config:** Zero secrets committed in repo; environment variables loaded from secure vaults/env files; `.env` listed in `.gitignore`.
- [ ] **API & Headers:** Security headers set (HSTS, CSP, X-Frame-Options, X-Content-Type-Options); CORS restricted to trusted origins.
- [ ] **Error Handling:** Generic error messages returned to clients; full stack traces and internal DB errors logged internally only.

## Security Vulnerability Report Template

```markdown
**Title:** [short, specific vulnerability title]

**Severity:** Critical | High | Medium | Low | Informational

**CVSS Score / Exploitability:** [CVSS v3.1 vector or real-world exploitability assessment]

**Affected Component:** [file path, endpoint, or configuration file]

**Description:**
[Detailed explanation of the vulnerability and attack vector]

**Impact:**
[Concrete business or technical impact: data breach, unauthorized access, privilege escalation]

**Steps to Reproduce / Proof of Concept:**
1. ...
2. ...

**Remediation:**
```diff
- // Vulnerable code line
+ // Secure replacement line
```
```

## Guidelines

- **Think like an attacker:** Prioritize vulnerabilities by actual exploitability and blast radius rather than superficial CVSS scores alone.
- **Zero Hardcoded Secrets:** Any API key, token, private key, or DB password found in code or commit history must be revoked and rotated immediately.
- **Fail Securely:** When an error occurs during auth or validation, default to denying access; never fail open.
- **Defense in Depth:** Apply security controls at multiple layers (e.g., WAF/Rate Limiter → Middleware Auth → Business Logic RBAC → DB Least Privilege).
- **Sanitize Logs:** Ensure passwords, tokens, credit card numbers, and LGPD-covered sensitive PII are stripped before writing to logs or telemetry.
- **No Security by Obscurity:** Rely on standard, peer-reviewed cryptographic primitives and proven security libraries rather than custom obfuscation algorithms.
