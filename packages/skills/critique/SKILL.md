---
name: critique
description: 'Adversarial implementation audit. Questions whether the current code implementation is reasonable, over-engineered, or unnecessarily complex. Triggers on: "critique", "question implementation", "是否合理", "合理性", "质疑实现", "是不是写太复杂了", "过度设计".'
---

# Critique (Implementation Sanity & Simplicity Audit)

Critique the current implementation approach. The goal is to act as a harsh critic focusing on simplicity, YAGNI (You Aren't Gonna Need It), robustness, and elegance. This audit happens _after_ coding is complete but _before_ formal code-review.

## Core Directives

1.  **Be Brutal on Over-Engineering**: Flag any extra layers, generic classes, or future-proofing code that is not needed for the current requirement.
2.  **Challenge Complexity (KISS)**: Ask "Can we do this in half the code?" or "Is this abstraction hiding simple logic?".
3.  **Audit Robustness**: Look for silent error ignores, race conditions, missing timeout settings, or unhandled null/error cases.

## Checklist

Perform the following audits on the changed files:

### 1. The YAGNI Check

- Did we write code for "future features"?
- Are there new configurations, methods, or classes that are never called or read?
- _Action_: Suggest deleting unused code immediately.

### 2. The Simplicity Audit (KISS)

- Could this be solved with a standard library function or simpler structure?
- Is the control flow convoluted (e.g., deep nesting, excessive callbacks, over-layered abstractions)?
- _Action_: Propose a concrete refactored alternative showing side-by-side simplification.

### 3. The Technical Debt Check

- Did we leave any temporary hacks, hardcoded constants, or `TODO` comments?
- Are we ignoring errors or logging them without handling?
- _Action_: List items that must be resolved before merging.

### 4. The Edge-Case Check

- How does this code behave if the database is down, network drops, or inputs are empty/malformed?
- Are concurrent operations safe?

## Output Format

Present the audit results clearly:

- **Verdict**: [Reasonable / Over-engineered / Over-complex]
- **Concerns**: Group issues by impact (High Complexity, YAGNI Violation, Robustness Risk).
- **Proposed Refactoring**: Provide a clear diff or code block showing a simpler implementation.
