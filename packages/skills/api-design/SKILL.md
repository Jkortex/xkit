---
name: api
description: 'Design external API contracts (REST, GraphQL, RPC) that are stable, consistent, and hard to misuse. Triggers on: "design API", "API contract", "REST endpoint", "define API", "接口设计", "API 设计".'
---

# Api

Design the external-facing contract. Call this from `spec` Phase 1 when the solution involves APIs, or standalone when adding/changing endpoints.

## When to use / skip

- **Use:** new endpoint, changing existing API, defining client-server contract
- **Skip:** internal-only module calls (use `interface` instead), one-off scripts

## Process

### 1. Resource model

Map domain entities to resources. Each resource gets one URL pattern.

```
POST   /todos           → create
GET    /todos           → list (with pagination/filter)
GET    /todos/:id       → get one
PATCH  /todos/:id       → partial update
DELETE /todos/:id       → delete
```

Consistent naming: plural nouns, kebab-case, no verbs in URLs.

### 2. Request / response contract

For each endpoint, specify:

- **Request shape** — path params, query params, body schema
- **Response shape** — success body, status code
- **Error modes** — what can go wrong, error shape, status codes (don't just return 500 for everything)
- **Timestamp format** — ISO 8601 with explicit timezone (`2026-05-31T15:00:00Z`). No mixed formats, no locale-dependent strings.
- **Data exposure** — flag every field that is PII, secret, or internal-only. If you wouldn't put it in a public doc, question why it's in the response.

### 3. Consistency checks

- Same pagination pattern across all list endpoints
- Same error shape across all endpoints
- Same naming convention for all fields (camelCase JSON, snake_case DB)
- Boolean fields use positive naming (`is_active`, not `inactive`)

### 4. Stability check

Apply Hyrum's Law: assume every observable behavior will be depended on. If you wouldn't document it, don't expose it. If you expose it, you own it forever.

---

Related: `module-interface` for internal module boundaries, `arch-depth` for retrospective architecture improvement.
