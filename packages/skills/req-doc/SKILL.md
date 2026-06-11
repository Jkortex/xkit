---
name: req-doc
description: 'Turn fuzzy ideas into a structured requirements document. Interview the user to clarify intent, then write a human-readable spec. Triggers on: "requirement", "需求", "我要做一个", "帮我理一下需求", "我想实现", "need a feature".'
---

# Req-doc

Translate fuzzy ideas into a clear requirements document. This is the human-facing input to the pipeline: write it → `grill-me` → `spec-to-plan`.

## When to use / skip

- **Use:** new feature request, vague idea that needs shaping, stakeholder handoff, feature request from non-technical source
- **Skip:** requirements are already clear and documented, one-line bug fix, technical refactor with no user-facing change

## Process

### 1. Explore intent

Ask clarifying questions ONE AT A TIME. Don't dump a list. Common dimensions:

- **Analogy** — "有没有跟哪个现有功能/产品很像的？" (e.g. "像 GitHub PR review 流程")
- **Who** — who uses this? What's their goal?
- **What** — what should happen? What should NOT happen?
- **Why** — what problem does this solve? What's the pain?
- **Where** — frontend? API? CLI? All of the above?
- **Priority** — must-have vs nice-to-have?

**When user is vague or stuck:** provide 2-3 concrete options rather than open-ended questions.

```
User: "我想加个通知功能"
AI: "您更想要哪种？A) 邮件实时通知  B) 系统内红点提示  C) 都可以"
```

Stop when you can describe the feature back to the user in one sentence and they agree.

### 2. Write req doc (in Chinese)

```
# 功能：<名称>

## 问题
<一句话描述痛点或机会>

## 目标用户
<谁受益，怎么受益>

## 需求
- R1: <必须实现，可验证>
- R2: ...
- Rn: <锦上添花>

## 非需求
<明确排除在范围之外的>

## 用户流程（可选）
<主流程步骤，步步有预期结果>
```

### 3. Present and approve

Show the doc. Ask: "Is this accurate? Anything missing?" Once approved, it's frozen.

---

Output feeds into: `grill-me` → `spec-to-plan`.
