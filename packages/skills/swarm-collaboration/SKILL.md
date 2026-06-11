---
name: swarm-collaboration
description: 'Delegate and execute tasks in parallel using a bundled Node.js script. Each sub-agent reports with a handoff summary. Triggers on: "parallel", "swarm", "delegate tasks", "并发执行", "并行任务".'
---

# Swarm Collaboration

Use the bundled [`scripts/parallel.mjs`](scripts/parallel.mjs) to execute multiple independent sub-tasks concurrently. Each sub-agent completes its task and outputs a structured **handoff summary** of what it accomplished.

## When to Use / Skip

- **Use when:**
  - Supplementing unit test coverage across multiple directories/packages.
  - Generating standard boilerplate code (CRUD, DTO mappings, ORM models) from schema designs.
  - Researching APIs/third-party libraries and compiling short summaries.
  - Making batch changes like lint fixing or documentation comment insertion.
- **Skip when:**
  - Designing core systems, architecture, or writing critical security-sensitive code.
  - Solving complex, stateful bugs that require deep reasoning across the entire codebase.
  - Making trivial, single-line modifications that don't benefit from parallelism.

## Process

### 1. Task Decomposition

Break down the request into multiple isolated, independent sub-tasks. Each task must:

- **Not modify files** that other tasks touch (avoid conflicts).
- Have a **clear, self-contained instruction** — the sub-agent should be able to complete it without external context.
- Include **context in the instruction** itself: which files to modify, what patterns to follow, what conventions to use.

### 2. Prepare `tasks.json`

Write a `tasks.json` file with the task definitions. The script reads this file.

```json
{
  "concurrency": 2,
  "tasks": [
    {
      "id": "write-user-tests",
      "instruction": "Write unit tests for user.go covering all functions...",
      "model": "opencode-go/deepseek-v4-flash",
      "thinking": "medium"
    },
    {
      "id": "fix-docs",
      "instruction": "Update README.md with new API endpoints...",
      "thinking": "low"
    }
  ]
}
```

Write it to `.xkit/tasks.json` (or any path):

```bash
mkdir -p .xkit
cat > .xkit/tasks.json << 'EOF'
{ ... }
EOF
```

### 3. Execute in Parallel

Find the skill directory, then run the bundled script:

```bash
SKILL_DIR=$(dirname "$(readlink -f ~/.agents/skills/swarm-collaboration 2>/dev/null || echo "")")
SKILL_DIR="${SKILL_DIR:-.agents/skills/swarm-collaboration}"
node "$SKILL_DIR/scripts/parallel.mjs" .xkit/tasks.json
```

The script will:
1. Spawn one `pi` process per task, up to `concurrency` at a time
2. Each sub-agent receives the instruction with an auto-appended handoff template
3. Wait for all tasks to complete
4. Print a summary to stdout
5. Exit 0 if all succeeded, 1 if any failed

The handoff files are written to `.agents/handoff-{task-id}.md`.

### 4. Review Handoffs

Read each sub-agent's handoff to understand what was accomplished:

```
---
### ✅ write-user-tests
## Handoff Summary
### 完成内容
- Wrote tests for CreateUser, GetUser, UpdateUser
### 修改的文件
- packages/users/user_test.go

---
### ✅ fix-docs
## Handoff Summary
### 完成内容
- Updated API reference in README.md
### 修改的文件
- README.md
```

For machine consumption, use `--json`:

```bash
node "$SKILL_DIR/scripts/parallel.mjs" .xkit/tasks.json --json
```

Returns a JSON array:
```json
[
  {
    "task_id": "write-user-tests",
    "status": "success",
    "handoff": "# Handoff: write-user-tests\n...",
    "duration_ms": 15234
  }
]
```

### 5. Handle Failures

If a task fails (non-zero exit and no handoff file produced), the script reports it with the error. Common causes:
- The instruction lacked context for the sub-agent
- The sub-agent hit a timeout (>5 minutes)
- The model was unavailable

**Retry strategy:**
- For transient failures (model busy), re-run the same task
- For logical failures (wrong file modified), improve the instruction and retry
- For persistent failures, fall back to sequential execution

## Schema Fields

| Field | Required | Description |
|-------|----------|-------------|
| `tasks` | ✅ | Array of task objects |
| `tasks[].id` | ✅ | Unique task identifier (used for handoff filename) |
| `tasks[].instruction` | ✅ | Instruction for the sub-agent |
| `tasks[].model` | ❌ | Override model (e.g. `opencode-go/deepseek-v4-flash`) |
| `tasks[].thinking` | ❌ | Thinking depth: `off`, `low`, `medium`, `high` |
| `tasks[].files` | ❌ | Files this task is expected to modify (informational) |
| `concurrency` | ❌ | Max parallel agents (default: 2) |

## Tips

- Keep tasks small and focused — a task should modify 1-3 files
- If two tasks would modify the same file, merge them into one task
- Use `--json` output for CI pipelines and automated tooling
- The handoff template is auto-appended; you don't need to include it in the instruction
