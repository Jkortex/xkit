---
name: grill-me
description: 'Adversarial plan/design review. Interrogates the proposed plan against the project''s architecture decisions (ADRs) and design conventions. Triggers on: "grill me", "prove me wrong", "what am I missing", "拷问", "忽略了", "怎么样".'
---

# Grill-me (Adversarial Design Review)

Grill the user on their plan or design. The goal is to play devil's advocate, expose hidden assumptions, and align the plan with the project's established conventions, architecture decision records (ADRs), and domain glossary.

## Process

### 1. Context Alignment (Monorepo Scoped)

Before asking questions, search both the global root and the target package directories for architecture rules:

- **ADRs**: Check global decisions under `docs/adr/*.md` and package-specific decisions under `packages/<pkg>/docs/adr/*.md` to ensure the plan does not violate previous architectural choices.
- **Design Specs & Context**: Check existing specifications and design documents at the root level (`docs/specs/*.md`) and within package subfolders (`packages/<pkg>/docs/**/*.md` or `packages/<pkg>/README.md`) to maintain proper boundaries.
- **Glossary / Domain Model**: Verify terminology against global and package-specific glossary/context files.

### 2. Systematic Challenge

Challenge the plan branch-by-branch. Focus on:

- **Deadzones**: What edge cases (e.g. offline, rate limits, concurrent runs, error states) are ignored?
- **Complexity**: Can we achieve 80% of the value with 20% of the complexity?
- **Coupling**: Does this change violate clean module boundaries?

### 3. Issue Severity Classification

Flag findings in a structured list:

- 🔴 **Must fix** — Violates an ADR, introduces a security flaw, or contains logical contradictions.
- 🟡 **Should discuss** — High risk, missing edge case, or questionable trade-off.
- 🔵 **Note** — Worth documenting, minor optimization, or assumption to verify.

For each flagged issue, propose a concrete, actionable alternative approach.

### 4. Interactive Resolution

Work through issues sequentially or in logical blocks. Stop only when all 🔴 and 🟡 issues are resolved and the user explicitly signs off.
