---
name: tdd
description: 'Red-green-refactor cycle (TDD). Focuses on behavior over implementation details, avoiding horizontal slicing and fragile mocks. Triggers on: "TDD", "red-green", "test first", "write test", "测试先行", "先写测试".'
---

# TDD (Test-Driven Development)

TDD is about confidence, design feedback, and living specifications. One cycle = one failing test ➔ one passing implementation ➔ refactor while green.

## Philosophy: Behavior over Implementation

- **Core principle**: Tests must verify behavior through public interfaces, not implementation details. Code can change entirely; robust tests shouldn't.
- **Good tests** are integration-style: they exercise real code paths through public APIs. They read like specifications (e.g. "user can checkout with valid cart") and survive refactoring because they don't care about internal file/function structure.
- **Bad tests** are tightly coupled to implementation: mocking internal collaborators, testing private methods, or verifying through external side-channels (like querying a database directly instead of checking the API). Fragile tests break when you rename an internal function or refactor, even when user-facing behavior is unchanged.

## Anti-Pattern: Horizontal Slices

- **DO NOT write all tests first, then all implementation.** Treating RED as "write all tests" and GREEN as "write all code" is a major anti-pattern.
- Writing tests in bulk results in testing _imagined_ behavior and the _shape_ of things (data structures, signatures) rather than real usage. You outrun your headlights and lock yourself into poor designs.
- **Correct Approach**: Vertical slices via tracer bullets. One small test ➔ one minimal implementation ➔ refactor ➔ repeat. Adjust the next test based on what you learned in the current cycle.

## The TDD Process

### 1. Red: Write a failing test

Write **ONE** test that describes the desired behavior.

- Run it using a **targeted, single-test command** (e.g. `npx vitest run TodoService.test.ts --reporter=verbose`), never the full suite.
- **Speed matters**: Feedback must be under 2 seconds. If it's slower, the test setup is too heavy or the command is too broad.
- Confirm it fails for the **RIGHT reason** (not syntax/compilation error, but an assertion failure proving the feature doesn't exist yet). If you didn't watch it fail, you don't know if it tests the right thing.

### 2. Green: Minimal code to pass

Write the simplest code that makes the test pass.

- No extra features. No future-proofing. No "while we're here." The goal is simply to get back to green as fast as possible.

### 3. Refactor: Clean up

Improve the code structure while keeping tests green.

- Extract duplication, improve naming, simplify complexity.
- Since behavior hasn't changed, tests are your safety net — if they stay green, your refactor is successful.

### 4. Loop

Repeat until the feature/bugfix is complete. Each cycle should take 2-10 minutes. If a cycle takes longer, your test step is too large — split it.

---

Called from `spec-to-plan` Phase 2 during incremental execution, or standalone for any feature/bugfix.
