---
name: arch-depth
description: 'Find shallow modules and propose deepening refactors. Use when code feels hard to test, hard to navigate, or every change touches too many files. Triggers on: "improve architecture", "refactor", "deepen this", "shallow module", "架构优化", "重构", "模块太浅".'
---

# Depth

Analyze existing code for **shallow modules** — interfaces as complex as their implementation, no information hiding, hard to test. Propose refactors to turn them into deep ones.

Run this before `interface` on existing code: find the problems first, then redesign.

## When to use / skip

- **Use:** code is hard to test, change propagates everywhere, module does too little, too many dependencies
- **Skip:** greenfield design (use `interface` instead), one-off scripts

## Process

### 1. Discover shallow modules

Scan for symptoms of shallow modules:

- Interface is as long as the implementation (pass-through)
- Callers need to understand internals to use it
- Changing one thing requires touching many files
- Module has too many public functions for what it does
- No clear seam for testing

### 2. Diagnose root cause

For each candidate, identify what makes it shallow:

- **Leaky abstraction** — exposes underlying mechanism instead of hiding it
- **Mixed concerns** — module does more than one thing, can't name its single responsibility
- **Layer pancake** — A → B → C with each adding nothing, just forwarding
- **God module** — too many responsibilities, interface is huge

### 3. Propose a deeper design

Generate alternative interface(s). Consider radically different approaches — parallel subagents can help here.

For each alternative, specify:

- New interface (smaller, hides more)
- What moves from interface to implementation
- How seams improve for testing
- Tradefoffs: what gets harder, what gets easier

**Before vs After matrix:** compare cognitive load — number of public functions, types, states, ordering requirements the caller must know.

```
                   Before        After
Public methods      4             1
Internal states     3             0
that leak out
Ordering rules      2             0
(call A before B)
```

The goal: fewer things a caller needs to hold in their head.

### 4. Validate

- Does the new design pass `interface` skill checks?
- Run `grill-me` on high-risk proposals

### 5. Document

Write a brief ADR (`docs/adr/`) covering: what was shallow, proposed new interface, tradefoffs, decision.

---

Related: `module-interface` for building new module boundaries, `spec-to-plan` for upfront design.
