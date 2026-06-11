---
name: code-cleanup
description: 'Simplify code without changing behavior. Reduce complexity, remove dead weight, make the next reader''s job easier. Triggers on: "simplify", "clean up", "refactor", "复杂", "简化", "清理".'
---

# Code-cleanup

Simplify code that works but is harder than it needs to be. Behavior must remain identical — tests are your safety net.

## When to use / skip

- **Use:** code is hard to follow, function too long, duplicate logic, confusing names, dead code
- **Skip:** before tests pass, during a bug hunt (fix first, clean later), when the code is about to be rewritten anyway

## Process

### 1. Scan for smells

- **Dead code** — unused functions, unreachable branches, commented-out code → delete
- **Duplication** — same logic in two places → extract once
- **Long function** — >20 lines doing multiple things → split by responsibility
- **Deep nesting** — 3+ levels of if/for → early return or guard clause
- **Complex conditional** — boolean soup → extract to named function
- **Variable pollution** — too many intermediate temps scattered through a method, forcing mental tracking → inline trivial ones, or extract a group of related vars + their operations into a Value Object via `module-interface`

### 2. Simplify

One change at a time. After each:

1. Run tests → confirm green
2. If not green, revert and try a smaller change

### 3. Verify

No behavior change. If a test fails, the simplification wasn't safe — revert.

---

Rule of thumb: after cleaning, the code should be **shorter or clearer**. If it's neither, revert and keep the original.
