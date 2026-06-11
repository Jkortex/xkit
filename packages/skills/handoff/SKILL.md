---
name: handoff
description: 'Compacts the current session''s context into a structured handoff document. Use when switching sessions, spawning a new subagent, or pausing work to be resumed later. Triggers on: "handoff", "session transfer", "交接", "工作交接", "保存会话", "生成交接文档".'
---

# Handoff (Session Context Compaction & Transfer)

Generate a structured handoff document to transfer development state between AI sessions or agents without losing critical context.

## When to Use

- **Session Token Exhaustion**: When the context window is near full and you need to start a fresh clean session.
- **Agent Handover**: When transitioning from a planning agent to a coding agent, or invoking a subagent.
- **Work Interruption**: When pausing work to be resumed later by another run.

## Process

1.  **Collect Status**: Identify what has been implemented, what is currently in progress, and what remains.
2.  **Map Artifacts**: List the spec documents, plans, or ADRs created in this session. Reference their file paths rather than duplicating their text.
3.  **Identify Blockers**: Highlight any unresolved bugs, open questions, or decisions waiting for user feedback.
4.  **Redact Secrets**: Ensure no api keys, passwords, or personal credentials are in the output.
5.  **Define Next Steps**: Write a concrete list of next actions, and recommend which skills (e.g., `tdd`, `code-review`) the next agent should run first.
6.  **Save File**: Write the handoff details to `.agents/handoff.md` (or `docs/handoffs/<date>-<topic>.md`).

## Handoff Document Template

```markdown
# Session Handoff: <Topic>

## 1. Context & Goal

- **Objective**: Brief explanation of what we are building.
- **Specs/ADRs**: [Title](file:///path/to/doc.md) (Reference paths, do not copy).

## 2. Current State

- **Completed**:
  - [x] Task A
  - [x] Task B
- **In Progress**:
  - [ ] Task C (Current file: [filename](file:///path/to/file.ts#L45))
- **Git State**: Branch `<branch-name>`, last commit `<SHA>`.

## 3. Open Issues & Blockers

- Unresolved decisions or bugs discovered.

## 4. Next Actions & Recommended Skills

1.  **Action 1**: ... ➔ _Invoke skill `tdd`_
2.  **Action 2**: ... ➔ _Invoke skill `code-review`_
```
