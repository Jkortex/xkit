# 架构决策记录（ADR）

## ADR-001：Memo 原子更新

**Status**: Accepted | **Date**: 2026-03-03

内容与标签单事务更新，避免"内容成功、标签未同步"窗口期。仓储层 `UpdateContentAndTags` 一次提交，含孤儿标签清理。

## ADR-002：导出先写临时文件再回传

**Status**: Accepted | **Date**: 2026-03-03

流式导出可能返回半截损坏包。改用临时文件完整生成 ZIP，成功后再响应。失败返回明确错误，不产生损坏文件。

## ADR-003：导入幂等多键去重

**Status**: Accepted | **Date**: 2026-03-03

资源按 `id/hash/internal_path`，Memo 按 `id/fingerprint(content+created_at)` 多键去重。同一归档重复导入后 imported 计数收敛为 0。

## ADR-004：错误语义标准化

**Status**: Accepted | **Date**: 2026-03-03

引入 `ErrInvalidInput`、`ErrNotFound` 等应用错误类型，handler 统一映射为 HTTP 400/404/500。调用方可区分可修复输入错误与系统错误。

## ADR-005：导入返回结构化报告

**Status**: Accepted | **Date**: 2026-03-03

导入响应包含 `report.memos/resources.details` 明细，展示 `duplicate_by_*`、`invalid_metadata` 等跳过原因。迁移可解释性提升。

## ADR-006：检索能力增强

**Status**: Accepted (Delivered) | **Date**: 2026-03-03

`GET /memos` 扩充 `from/to/has_resource/tags_any/tags_all/sort` 参数。保持向后兼容，新增查询索引 `(row_status, created_at)`、`(row_status, updated_at)`。

## ADR-007：键盘优先交互

**Status**: Accepted (Delivered) | **Date**: 2026-03-03

渐进增强：保留鼠标入口，新增全局快捷键（`/`, `Ctrl+K`, `N`, `J/K`, `E`, `D`, `?`, `Esc`）+ 组合筛选控件。

## ADR-008：可视化导入报告

**Status**: Accepted (Delivered) | **Date**: 2026-03-03

导入成功后前端弹出结构化报告弹窗，展示导入/跳过计数与原因明细。无需新增后端接口。

## ADR-009：标签重命名即合并

**Status**: Accepted (Phase 1/2 Delivered) | **Date**: 2026-03-03

`POST /tags/rename`: 目标不存在→重命名，已存在→自动合并。Phase 2 新增 `POST /tags/merge` 批量合并。

## ADR-010：标签别名规范化

**Status**: Accepted (Delivered) | **Date**: 2026-03-03

`tag_alias(alias → canonical)` 映射。写入/查询路径自动解析。限制解析深度防循环。

## ADR-011：标签治理审计流

**Status**: Accepted (Delivered) | **Date**: 2026-03-03

`tag_governance_audit` 表记录 rename/merge/alias 操作。`GET /tags/audits` 可查。

## ADR-012：全面采用 UUIDv7

**Status**: Accepted (Delivered) | **Date**: 2026-03-10

MemoUUID/Resource/Invite/Session 全部升级 UUIDv7。时间有序性减少 SQLite B-Tree 页分裂，提升大批量写入性能。导入幂等基于 UUIDv7 主键。

## ADR-013：业务数据按用户隔离

**Status**: Accepted (Delivered) | **Date**: 2026-03-08

`owner_user_id` 字段 + 读写链路按当前用户作用域执行。API 显式传递 `userID`，不采用隐式上下文注入。

## ADR-014：不采用乐观更新

**Status**: Accepted | **Date**: 2026-03-10

所有写操作遵循 API → 成功响应 → Store 更新 → 反馈用户链路。理由：笔记数据安全优先级 > 交互速度；乐观更新使 Pinia Store 膨胀 30-50%；实际写入延迟 < 50ms，收益不足以抵消风险。

## ADR-015：编辑历史与版本控制

**Status**: Accepted | **Date**: 2026-03-10

`memo_history` 表全量快照（非 Diff），仅真实变化时存档。单笔记保留最近 20 条。GC 需检查 `memo_history` 引用，防止回滚后资源丢失。前端实时计算 Diff。

## ADR-016：以键盘为中心资源联动

**Status**: Accepted | **Date**: 2026-03-10

`![描述](res-UUID)` 语法引用内部 Resource ID。`Ctrl+U` 上传后自动插入引用。渲染层将 `res-` 引用解析为真实 API 地址。

## ADR-017：外链视频 + 10MB 限制

**Status**: Accepted | **Date**: 2026-03-10

本地坚持 10MB 上限。`![视频](url.mp4)` 渲染为 `<video>`。11MB+ 前端的拦截提示。

## ADR-018：末尾块标签提取

**Status**: Accepted | **Date**: 2026-03-13

自底向上扫描，纯标签行提取为正式标签。正文中 `#Tag` 不被提取。容忍末尾空行。

## ADR-019：导入路径校验

**Status**: Accepted | **Date**: 2026-03-13

`isValidInternalPath`: `filepath.Clean` 规范化，拒绝绝对路径和 `../`。防止目录穿越攻击。

## ADR-020：修正 SQLite 批量操作 ON CONFLICT 兼容性与清理 CLI stale references

**Status**: Accepted | **Date**: 2026-06-05

1. **SQLite 批处理优化**: 修正了 `BatchSaveTags` 和 `BatchTagAdd` 在 SQLite 上的语法兼容性。将 PostgreSQL 风格的 `ON CONFLICT DO NOTHING` 重构为 SQLite 原生的 `INSERT OR IGNORE INTO`，避免了由于缺少 conflict target 在部分 SQLite 驱动中引发的 `DO` 附近语法错误。
2. **清理 stale references**: 修复了 `packages/xkit-cli` 自动构建时的性能漏洞。移除了已被物理删除的 `packages/workspace` 的 `needsRebuild` 检查，防止其因目录不存在导致每次调用 CLI 工具均发生重复全量编译。
3. **补全测试脚手架**: 补全了 `tests/integration/test_utils.go` 中缺失的 `setupTestApp` 与 `newTestClient` 测试脚手架实现，打通了 `batch_api_test.go` 的编译与集成测试链路。
