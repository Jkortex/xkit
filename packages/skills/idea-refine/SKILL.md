---
name: idea-refine
description: 'Refine fuzzy ideas through structured divergent and convergent thinking. Use before req-doc when the concept isn''t clear yet. Triggers on: "idea", "brainstorm", "ideate", "想法", "点子", "头脑风暴".'
---

# Idea-refine

Turn a vague notion into a concrete proposal worth specifying. Two phases: widen, then narrow.

## When to use / skip

- **Use:** you have a hunch but can't articulate it, exploring approaches before committing, "wouldn't it be cool if..."
- **Skip:** requirements already clear (go straight to `req-doc`), trivial decisions (use `grill-me` instead)

## Process

### 1. Diverge (widen)

Generate options without judgment. Quantity over quality.

Ask from different angles:

- **Alternatives** — what are 3-5 different ways to solve this?
- **Opposite** — what if we did the opposite?
- **Remove constraint** — if X weren't a limitation, what would we do?
- **Analogies** — how do other products/tools solve this? (see `req-doc` analogy dimension)

Collect all ideas in a flat list. No critiquing yet.

### 2. Converge (narrow)

Evaluate and filter:

- **Must-have criteria** — what's non-negotiable? Discard options that fail these.
- **Tradeoffs** — for each remaining option, what's the cost, complexity, and risk?
- **Pick one** — recommend the best approach and state why.

### 3. Output

A one-paragraph description of the refined idea, ready to feed into `req-doc`:

```
What: <core idea in one sentence>
Why: <problem it solves>
How (rough): <high-level approach>
Why this over others: <key tradeoff that decided it>
```

---

Output feeds into: `req-doc` → `grill-me` → `spec-to-plan`.
