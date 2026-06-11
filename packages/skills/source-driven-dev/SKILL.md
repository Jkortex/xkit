---
name: source-driven-dev
description: 'Ground every implementation decision in official documentation. Never guess APIs — read the source. Triggers on: "source driven", "check docs", "official docs", "according to spec", "查文档", "官方文档".'
---

# Source-driven-dev

Code based on memory or blog posts is code with hidden bugs. Before writing code that touches a framework, library, or API, read the official docs.

## When to use / skip

- **Use:** using an unfamiliar API, implementing against a spec, integrating a library, choosing between approaches
- **Skip:** trivial (language built-ins you use daily), self-contained business logic with no external API

## Process

### 1. Read the source

Before writing code against any external API:

- **检查会话历史（避免会话内重复获取）**：如果当前会话前文中已经搜索、读取或拉取过相关的官方文档或 API 规范，**请直接复用历史上下文中的内容，严禁重复发起网络请求（如重复搜索或重新读取相同 URL）**。
- **检查本地文档缓存（避免跨会话重复获取）**：
  - 在发起任何网络搜索或读取外部 URL 前，**必须先检查项目本地 [cache](file:///c:/Users/light/code/xkit/packages/skills/cache/) 目录**是否已存在对应库或 API 的文档缓存（如 `packages/skills/cache/<lib-name>.md`）。
  - 如果存在本地缓存，**直接使用工具读取本地缓存文件**，实现 100% 跨会话免网检索。
  - 如果不存在，在首次通过网络成功获取文档后，**应当主动将其整理并写入本地 `packages/skills/cache/<lib-name>.md` 中**，为未来的会话（以及其他协同 Agent）留下持久化的本地缓存。
- 寻找**官方文档**（拒绝博客或 Stack Overflow 等二传手内容）
- 寻找**真实的类型签名**（TypeScript 类型定义、Go doc、OpenAPI 描述文件等）
- 阅读**错误与边界处理章节**——输入无效时会发生什么？

**If you're guessing what a function returns, you're doing it wrong.**

### 2. Cite your source

When using a non-obvious API pattern, add a comment referencing the source:

```ts
// https://api.example.com/docs/pagination#cursor-based
const cursor = await db.query({ limit, after: cursorParam });
```

This saves the next reader (and future you) from having to re-search.

### 3. Verify

After writing, check:

- Did the docs say anything about error modes I'm not handling?
- Is there a newer/different API I should be using instead?
- Did I miss a deprecation notice?

---

Used during `spec-to-plan` Phase 1 (design) and `tdd` (implementation).
