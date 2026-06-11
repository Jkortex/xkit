---
name: security-review
description: 'Security review — find vulnerabilities before they ship. Treat every input as hostile, every secret as sacred. Triggers on: "security", "harden", "vulnerability", "安全", "漏洞", "加固".'
---

# Security-review

Security isn't a phase — it's a constraint on every line that touches user data, auth, or external systems.

## When to use / skip

- **Use:** any feature handling user input, auth, PII, payments, file upload, webhooks, external API integration
- **Skip:** internal-only CLI tools with no external input, pure computation with no I/O

## Checklist

### Input

- [ ] All external input validated (type, length, range, format) — never trust the client
- [ ] No raw string interpolation in SQL / shell / HTML — use parameterized queries, builders, or escaping
- [ ] File uploads: size limit, type whitelist, no direct path traversal

### Auth & access

- [ ] Auth check at every endpoint, not just in the UI
- [ ] **IDOR check** — does the endpoint verify the authenticated user OWNS the resource identified by the request? AuthN ≠ AuthZ. Passing `id=100` while logged in is NOT enough; confirm `100` belongs to the caller.
- [ ] No hardcoded credentials, tokens, or API keys in code
- [ ] Rate limiting considered for public endpoints

### Data

- [ ] No PII/secrets in logs, error messages, or URLs
- [ ] Sensitive fields flagged in API response (see `api-design` data exposure check)
- [ ] Encryption in transit (TLS) and at rest where needed

### Dependencies

- [ ] No known-vulnerable dependencies (check lockfile)
- [ ] No vendored code from untrusted sources

---

Run during `code-review` for security-sensitive changes, or standalone for a full security audit.
