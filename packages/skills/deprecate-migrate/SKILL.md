---
name: deprecate-migrate
description: 'Safely remove old code, APIs, or systems. Deprecate, migrate, clean up — never break callers. Triggers on: "deprecate", "migrate", "remove", "sunset", "弃用", "迁移", "下线".'
---

# Deprecate-migrate

Removing code is harder than writing it. The goal: callers never break, data never lost, old code eventually dies.

## When to use / skip

- **Use:** removing a public API, replacing a library, decommissioning a service, renaming a module
- **Skip:** internal code with no callers (just delete it with `code-cleanup`)

## Process

### 1. Find all callers

Before touching anything:

- Code search all usages of the API/module being removed
- Check for transitive callers (callers of callers)
- Logs/analytics — is anyone actually using this in production?

If you don't know who calls it, you can't deprecate safely.

### 2. Deprecation period

- Mark the old API as deprecated with a clear message: `@deprecated use X instead since v1.2`
- Keep it working during the migration period
- If external: announce deprecation timeline publicly

### 3. Dual-write (for data migrations)

When changing data format or storage:

- Write to both old and new locations simultaneously
- Compare results to verify correctness
- Backfill existing data before switching reads
- Switch reads to new location
- Stop writing to old location
- Remove old location

### 4. Remove

Only remove after:

- All callers migrated (confirmed by code search)
- Deprecation period elapsed
- No production errors on the new path
- Rollback plan ready (keep the old code in git, tag the removal commit)

---

The safest deprecation: one that nobody notices.
