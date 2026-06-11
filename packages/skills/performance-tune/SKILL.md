---
name: performance-tune
description: 'Measure, identify bottlenecks, optimize, verify. Never optimize without a baseline. Triggers on: "performance", "slow", "optimize", "bottleneck", "性能", "优化", "慢".'
---

# Performance-tune

Optimization without measurement is guessing. Every change starts and ends with a number.

## When to use / skip

- **Use:** slow page load, high latency API, janky UI, database queries getting slower, memory growth, bundle too large
- **Skip:** premature optimization during feature work (finish it first), micro-optimizations with no measurable impact

## Process

### 1. Establish baseline

Measure before touching anything:

- **Latency** — p50, p95, p99 response times
- **Throughput** — requests per second, queries per second
- **Resources** — CPU, memory, network, disk
- **Client** — LCP, FID, CLS (Web Vitals), bundle size

Without a number, you don't know if you fixed it.

### 2. Identify bottleneck

Find where time is actually spent, not where you think it is:

- **Profiling** — CPU flamegraph, allocation profile (not guesswork)
- **Tracing** — distributed trace for multi-service calls
- **Query analysis** — `EXPLAIN ANALYZE`, slow query log, N+1 detection
- **Bundle analysis** — source-map explorer, import visualization

### 3. Optimize

Target the bottleneck, not everything around it:

| Pattern               | Approach                               |
| --------------------- | -------------------------------------- |
| N+1 queries           | Eager loading, dataloader, batch       |
| Slow query            | Index, denormalize, materialized view  |
| Large bundle          | Code-split, tree-shake, dynamic import |
| Re-render             | Memo, selector, virtual scroll         |
| Blocking I/O          | Async, connection pool, cache          |
| Expensive computation | Memoization, lazy eval, worker         |

### 4. Verify

Re-run the same measurement from step 1. If it's not measurably better, revert.

**Diminishing returns rule:** if a second optimization attempt yields <10% improvement, stop. The remaining gain isn't worth the complexity.

---

Save baseline and result numbers alongside the change (commit message or issue comment).
