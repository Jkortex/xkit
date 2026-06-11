# 基于 Hook 的声明式自动留档协议 (DAP)

## 1. 背景与目标

在 Agent 协作流中，长文档（如计划、调研报告）的重复传输会迅速消耗 Token 额度。本项目旨在通过 Gemini CLI 的 `AfterTool` 钩子系统，实现“写文件即备份”的静默留档机制，在不增加额外对话轮次的情况下，确保 Agent 的产出具有永久可追溯性。

## 2. Agent 协议 (Skill 定义)

Agent 在生成需要留档的内容时，必须遵循以下 **DAP 协议**：

### 2.1 触发机制

Agent 使用 `write_file` 工具写入文件，且文件内容末尾必须包含特定的**声明标签块**。

### 2.2 标签规范

- `#sync`: **强制触发器**。标记该文件需要同步至 Daily 云端。
- 业务标签 (可选): 如 `#plan`, `#task`, `#research`, `#rfc` 等，用于分类。
- 关联标签 (可选): 如 `#p:<plan_name>` 用于关联特定计划。
- 溯源标签 (可选): 如 `#by:agent`。

### 2.3 示例格式

```markdown
# 架构设计方案 v1.0

（正文内容...）

#arch #rfc #by:agent #sync
```

## 3. 系统实现 (Hook 逻辑)

系统监听 `write_file` 工具的 `AfterTool` 事件，并执行以下自动化逻辑：

### 3.1 拦截与解析

1. **匹配工具**：仅拦截 `write_file` 及其变体。
2. **提取内容**：从工具输入中获取 `path` 和 `content`。
3. **扫描触发器**：检查 `content` 是否包含 `#sync`。若无，则静默退出。

### 3.2 数据处理与推送

1. **内容清洗**：
   - 提取所有以 `#` 开头的标签。
   - **移除 `#sync` 标签**，确保云端存储的是最终状态。
   - 保留其他所有标签。
2. **执行推送**：
   - 调用 `daily-cli post --file <path>`。
   - 优势：利用已落盘的文件，避免在 Shell 缓冲区中再次展开长文本。
3. **状态记录**：获取服务器返回的 `UUID`。

### 3.3 上下文注入 (Token 优化核心)

Hook 将执行结果作为 `additionalContext` 返回给 Gemini CLI：

- **返回示例**：`[System] DAP: Content in <path> archived to Daily (UUID: <uuid>)`。
- **效果**：Agent 在下一轮对话中获得该 UUID，无需重新读取已备份的内容。

## 4. 方案优势

1. **Token 极度节约**：长文档在对话上下文中仅出现一次（`write_file` 时），后续操作均为引用。
2. **非侵入式架构**：Agent 只需要“学会带标签写文件”，无需学习复杂的 API 调用。
3. **数据纯净性**：自动化 Hook 负责清理冗余的控制标签（如 `#sync`）。
4. **双重备份**：内容同时存在于本地文件系统（`docs/archived/`）和云端数据库（`Daily`）。

---

## 5. 未来实施清单

- [ ] 创建 `scripts/daily-sync-hook.js` 处理逻辑。
- [ ] 在 `.gemini/settings.json` 中配置 `AfterTool` 匹配器。
- [ ] 为 `daily-cli` 完善 `--file` 读取参数。
- [ ] 在主 `GEMINI.md` 中固化此 Skill 指令。
