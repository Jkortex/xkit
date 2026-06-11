---
name: interface
description: 'Design internal module boundaries — interfaces that are stable, testable, and hard to misuse. Triggers on: "module interface", "internal API", "component boundary", "模块接口", "解耦", "边界设计".'
---

# Interface

Design internal module boundaries. Call from `spec` Phase 1 when the solution involves multiple modules, or standalone when refactoring internal structure.

## When to use / skip

- **Use:** defining module boundaries, splitting a large module, improving testability, designing with multiple contributors
- **Skip:** single-file scripts, trivial wrappers, one-off internal utils

## Process

### 1. Module discovery

List the modules needed. For each, ask: what does it hide? A good module hides one thing completely.

```
Example: not "todo-service", but:
  todo-store     → hides persistence (SQL vs file vs in-memory)
  todo-policy    → hides business rules (who can edit what)
  todo-notifier  → hides delivery mechanism (email vs push vs none)
```

### 2. Interface definition per module

Define what callers MUST know (this IS the interface):

- Exported types / signatures
- Invariants (what's always true after a call)
- Error modes (what can fail)
- Ordering constraints (call A before B?)
- Config / construction (what do you need to create it)
- **Concurrency & lifecycle** — is the implementation thread-safe? Singleton, scoped, or transient? Flag if callers need to manage lifecycle (close, dispose).

### 3. Depth check

A deep module has a small interface and large implementation. A shallow module has an interface nearly as complex as its implementation.

For each module, ask:

- Does this interface hide complexity or just pass it through?
- Would callers need to understand the implementation anyway?
- Can we merge shallow adjacent modules into one deeper one?

**Cognitive load check:** count what callers MUST know — public functions, types, states, ordering rules. Fewer = deeper. If the count feels high, the interface is too shallow.

### 4. Seams

Where would you inject a test double? If you can't point to a seam, the design isn't testable. Add one — an interface, a callback, a parameter — at every external dependency.

### 5. Dependency isolation (ACL)

For each module, check: does it directly expose types/models from a third-party SDK or legacy system? If yes, you need an **anti-corruption layer** — translate at the boundary. External types must not leak past the module that owns the integration.

This keeps internal modules decoupled from vendor churn and makes them testable without real SDKs.

### 6. Stability check

Assume every public function will be called by five other modules within a year. Is this interface still a good one? If not, it's too leaky.

---

Related: `api-design` for external-facing API contracts, `arch-depth` for retrospective architecture deepening.
