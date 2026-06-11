---
name: spec-to-plan
description: 'Turn fuzzy requirements into a structured spec + incremental implementation plan. Triggers on: "write a spec", "plan this", "spec this out", "scope this", "break this down", "设计一下", "规划一下", "写个方案", "拆解", "先规划".'
---

# Spec-to-plan

Three phases: Context → Spec → Plan. Each has a gate — don't skip to next without sign-off.

## When to use / skip

- **Use:** new feature, ambiguous requirements, multi-module change
- **Skip:** single-line fixes, typos, trivial self-contained changes

## Phase 1: Context → Spec

1. **Explore codebase** — understand current state, relevant modules, ADRs. Use domain vocabulary.
2. **Synthesize** — do NOT interview. Work from what you know + exploration.
3. **Design solution** — iterate between `api` and `interface` until stable:

   a. Run **`api-design` skill** — define external-facing contracts (REST, GraphQL, etc.)
   b. Run **`module-interface` skill** — define internal module boundaries, seams, depth
   c. Iterate: API change may force module split; module split may simplify API
   d. Stable when: `module-interface` cognitive load check passes for every module (no shallow interfaces), and `api-design` has no unresolved consistency issues

   Output:
   - Problem / goal
   - Solution sketch — high-level approach, data flow between components
   - Module decomposition — with interfaces, seams, depth rationale
   - API / interface contract — signatures, types, invariants, error modes
   - Acceptance criteria — each with a verifiable check
   - Non-goals / out of scope

4. **Gate:** Present to user. Go → Phase 2. No-go → refine based on feedback, max 2 rounds, or kill.

## Phase 2: Spec → Plan

1. **Map file structure** — exact paths, clear boundaries.
2. **Break into incremental steps** — each step leaves the codebase in a working, testable state.
3. **Per step:** what to do, which files, test approach. Execute each step via `test-first` (red-green-refactor).
4. **Gate:** Present to user. Go → implement. No-go → refine or kill.

## Optional Phase 3: Validate

Run `grill-me` for high-risk changes. Direct sign-off for routine ones.

---

Save output to `docs/spec/<feature>.md` or project issue tracker.
