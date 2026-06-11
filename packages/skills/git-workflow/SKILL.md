---
name: git-workflow
description: 'Branch, commit, merge, rebase, and resolve conflicts. Keep history clean and traceable. Triggers on: "git", "commit", "branch", "merge", "rebase", "提交", "分支", "合并".'
---

# Git-workflow

Git discipline is team discipline. Clean history makes debugging, review, and rollback reliable.

## When to use / skip

- **Use:** before committing, when planning branches, resolving conflicts, preparing a PR
- **Skip:** single-file scratch changes you'll squash, exploratory branches no one else sees

## Process

### 1. Branch strategy

- **`main`** — production-ready. Merged only via PR with review
- **`feat/<name>`** — feature branch off `main`. Delete after merge
- **`fix/<name>`** — bugfix branch off `main`
- Avoid long-lived branches (>1 week). Split into smaller iterations.

### 2. Commit discipline

**When to commit:** after each green test from `tdd`. A commit should be a single logical change.

**Commit message:**

```
<type>(<scope>): <imperative summary>

<optional body: what and why, not how>
```

Types: `feat`, `fix`, `refactor`, `test`, `docs`, `perf`, `chore`

One commit = one concern. If you need "and" in the summary, split it.

### 3. Before push

Run `pre-push`. If anything is red, don't push.

### 4. Rebasing (not merging)

- Rebase feature branches onto `main` before PR, don't merge `main` into feature
- Interactive rebase to clean up commits before push (`git rebase -i`)
- Never rebase a branch others have pulled from

### 5. Conflict resolution

When conflicts arise during rebase:

1. Read both sides — understand the intent of each
2. Keep the better approach, not just "both"
3. After resolving, continue rebase and verify with tests

---

Related: `pre-push` for pre-commit checks, `tdd` for commit cadence.
