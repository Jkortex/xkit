---
name: code-review
description: 'Multi-axis code review before merge. Covers correctness, readability, architecture, security, performance. Triggers on: "review", "code review", "review this", "CR", "代码审查", "review一下".'
---

# Code-review

Multi-axis review. Every change must be reviewed before merging to ensure software quality and architectural alignment.

## Philosophy: Technical Rigor over Social Comfort

- **Review is evaluation, not performance**: Reviewing code requires objective technical inspection. Keep focus on the actual code changes, not the developer's thought process.
- **Verify before implementing**: When receiving feedback, check suggestions against codebase reality before writing code.
- **No performative agreement**: Avoid empty phrases like "You're absolutely right!" or "Great point!". Instead, state the technical requirements, ask clarifying questions, or push back with technical reasoning if the feedback is incorrect.

## Process Flow

1.  **Isolate Context**: Do not review based on raw session history. Obtain the precise changes using:
    ```bash
    git diff <BASE_SHA> <HEAD_SHA>
    ```
2.  **Evaluate across Five Axes**: Audit the diff against the review checklist below.
3.  **Categorize Findings**: Group issues by severity (🔴 Critical, 🟡 Important, 🔵 Minor).
4.  **Resolve & Implement**: Implement fixes one at a time, verifying each change via test runners. Clarify unclear feedback before proceeding.

## Issue Severity

- 🔴 **Critical (Must fix)** — Definite bugs, security holes, data loss risks, or violations of design specs/ADRs.
- 🟡 **Important (Fix before merge)** — Risky patterns, missing edge cases, code-smells, or coupling issues.
- 🔵 **Minor (Note)** — Style preferences, future opportunities, or minor refactors.

## Review Axes

### 1. Correctness

- Does the code meet all requirements and acceptance criteria in the spec?
- Are edge cases (empty states, errors, boundary values) handled?
- Are there tests covering the change? Do they verify behavior (not implementation detail)? (see `tdd`)
- **Resource leaks**: Are streams, DB handles, and locks safely released in `finally`/defer?
- **Concurrency safety**: Is shared state properly synchronized? No execution order assumptions?

### 2. Readability

- Can a new team member understand this code in one pass?
- Do variable and function names reveal intent, not implementation?
- **No magic literals**: Extract numbers and strings to named constants.
- No commented-out code blocks or dead code left in.

### 3. Architecture

- Does the change respect package and module boundaries? (see `module-interface`)
- Is it compliant with existing global and package-specific ADRs?
- No leaky abstractions or vendor type pollution.

### 4. Security

- Are all inputs validated? No raw interpolation in SQL or shell executions?
- Are authentication/authorization checks implemented in the correct layer?
- No secrets or credentials exposed in the diff. (see `security-review`)

### 5. Performance

- Are there N+1 queries, unnecessary allocations, or missing indexes?
- Does the change fit the system latency and memory budgets?

## Standard of Approval

**Approve when the change improves overall code health, even if it is not perfect.** Do not block on style preferences. Block only on unresolved 🔴 and 🟡 items.

---

Run after `tdd` completes, before merge. Handoff to `grill-me` for high-risk changes.
