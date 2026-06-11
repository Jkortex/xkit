---
name: pre-push
description: 'Local pre-push checklist. Verify everything is clean before you push — lint, typecheck, test, changelog, migration, review status. Triggers on: "ship", "pre-push", "push", "准备提交", "合并", "pre-commit".'
---

# Pre-push

What you can (and must) check locally before push. No CI, no CD — just your machine and your discipline.

## When to use / skip

- **Use:** before pushing any branch, before opening a PR, before merging
- **Skip:** WIP commits you'll squash later

## Checklist

- [ ] **Lint & typecheck (Monorepo scoped)** — run workspace-wide checks (e.g. `pnpm run lint`) or package-specific commands to verify changed packages. Zero errors across all affected areas.
- [ ] **Tests pass (Monorepo scoped)** — run full test suite or targeted package tests (e.g. `pnpm --filter <pkg> test`) via `tdd` to ensure local changes are fully covered and pass.
- [ ] **No 🔴 review items** — run `code-review` on your diff. All 🔴 resolved.
- [ ] **Changelog** — if user-facing change, entry added
- [ ] **Migrations** — if any, they're backward-compatible (no drop/rename without dual-write)
- [ ] **Secrets** — no keys, tokens, or credentials in the diff (use `security-review` for sensitive changes)
- [ ] **Dead code** — no commented-out blocks, no `console.log`/`debugger`/`TODO` left in (see `code-cleanup`)
- [ ] **Self-review** — read your own diff before asking anyone else to

---

Run before every push. If anything is red, don't push.
