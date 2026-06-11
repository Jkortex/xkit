---
name: debug-fix
description: 'Systematic debugging — find root cause before attempting any fix. Triggers on: "bug", "debug", "broken", "failing", "error", "不工作", "报错", "有问题", "排查".'
---

# Debug-fix

The Iron Law: **NO FIX WITHOUT ROOT CAUSE.** Symptom patching is failure.

## When to use / skip

- **Use:** any bug, test failure, unexpected behavior, performance regression
- **Skip:** feature work (use `tdd` instead), obvious typo fix (but confirm root cause first)

## Process

### Phase 1: Build a feedback loop

This is the most important phase. Without a fast, deterministic way to reproduce the bug, you can't find the cause.

Spend disproportionate effort here. Be aggressive. Be creative. Try in order:

1. **Can you reproduce it with a single command?** If not, write one.
2. **Minimise** — strip away everything not needed to trigger the bug (minimum input, minimum config, minimum code path)
3. **Isolate** — binary search the code path (comment out halves, git bisect)
4. **Instrument** — add logging at key decision points
5. **Deterministic signal** — the goal is a pass/fail signal that an agent can run in under 5 seconds

**If you can't reproduce it, you don't have a bug — you have a mystery.**

### Phase 2: Find root cause

Only after Phase 1 produces a reliable signal:

1. Read the code at the failure point
2. Trace backward: what inputs/state led here?
3. State your hypothesis explicitly (e.g. "the cache returns stale data because TTL isn't checked on read")
4. Test the hypothesis by running the feedback loop

### Phase 3: Fix

1. Write a **regression test** that fails with the bug and passes with the fix (`tdd`)
2. Implement minimal fix
3. Run the full feedback loop to confirm

### Phase 4: The 3-strike rule

If three separate fix attempts fail (the bug persists or a new one emerges), STOP. The problem isn't the code — it's the design. Run `arch-depth` to question the module structure.

---

Output: regression test + fix + brief postmortem (what was root cause, why it was missed).
