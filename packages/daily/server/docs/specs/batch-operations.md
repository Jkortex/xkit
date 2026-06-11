# Spec: Batch Operations for Memos

## Objective

当前所有 Memo 操作都是单条模式。用户进行标签治理后需要批量归档关联 Memo，或对导入数据批量打标时，只能逐条操作。本 spec 定义批量归档、批量删除、批量打标三个能力的 API 设计与实现方案。

**用户故事**：

- 作为日常用户，我想一次选择多条 Memo 并归档，而不是逐条打开删除
- 作为 CLI 用户，我想用一条命令批量打标或归档
- 作为治理用户，合并标签后我想一次归档所有受影响的不再需要的 Memo

## Tech Stack

- **Server**: Go 1.23+, Gin, sqlc (pgx/v5), SQLite (FTS5), PostgreSQL 16
- **Web**: Vue 3, TypeScript, TDesign Vue Next, Pinia
- **CLI**: Go, Cobra (in `server/cmd/daily-cli/`)

## Commands

```bash
# Server
cd packages/daily/server && make test-sl          # 单元+集成测试
cd packages/daily/server && make test-pg          # Postgres 测试
cd packages/daily/server && make doc-route-check   # 文档门禁

# Web
cd packages/daily/web && pnpm test                # Vitest

# CLI
cd packages/daily/server && go test -v ./cmd/daily-cli/...  # Go 测试
cd packages/daily/server && go build -o daily-cli ./cmd/daily-cli/  # 构建
```

## Project Structure

```
server/
├── internal/
│   ├── application/
│   │   ├── dto/                     ← BatchArchiveResponse, BatchTagResponse
│   │   ├── port/                    ← Repository 新增 BatchArchive, BatchTag 接口
│   │   └── usecase/memo/            ← NewUseCase: BatchArchive, BatchDelete, BatchTag
│   └── infrastructure/
│       ├── api/handler/             ← 3 个新 Handler: BatchArchive, BatchDelete, BatchTag
│       └── persistence/sqlite|postgres/  ← sqlc 生成批量 SQL + Repository 实现
├── migrations/                      ← 本次不需要迁移（无 schema 变更）

web/
├── src/
│   ├── presentation/components/     ← MemoList 新增多选模式 + BatchActionBar
│   ├── application/services/        ← BatchService
│   ├── infra/gateway/               ← HttpBatchGateway
│   └── application/ports/           ← IBatchPort, IBatchService

cli/
├── src/
│   ├── cli/args.rs                  ← 新增 batch archive/delete/tag 子命令
│   ├── cli/handlers.rs              ← handle_batch
│   ├── domain/ports.rs              ← 新增 BatchOps trait
│   ├── domain/models.rs             ← BatchRequest, BatchResponse
│   └── infra/api.rs                 ← ApiClient 实现 batch 调用
```

## Code Style

遵循各项目已有风格：

**Go (server):**

```go
// 批量归档 usecase
func (s *MemoService) BatchArchive(ctx context.Context, userID int64, uuids []string) (*dto.BatchResult, error) {
    if len(uuids) == 0 {
        return nil, fmt.Errorf("%w: uuids cannot be empty", apperr.ErrInvalidInput)
    }
    if len(uuids) > 100 {
        return nil, fmt.Errorf("%w: max 100 uuids per batch", apperr.ErrInvalidInput)
    }
    archived, err := s.repo.BatchArchive(ctx, userID, uuids)
    if err != nil {
        return nil, fmt.Errorf("batch archive: %w", err)
    }
    return &dto.BatchResult{Succeeded: archived}, nil
}
```

**TS (web):**

```typescript
// BatchService
async archiveBatch(uuids: string[]): Promise<Result<BatchResponse>> {
  if (uuids.length === 0) return success({ succeeded: [], failed: [] });
  if (uuids.length > 100) return failure(new Error('Max 100 memos per batch'));
  return this.gateway.batchArchive(uuids);
}
```

**Rust (CLI):**

```rust
// args.rs
#[derive(Subcommand)]
pub enum BatchCommands {
    Archive { ids: Vec<String> },
    Delete { ids: Vec<String> },
    Tag { ids: Vec<String>, add: Option<Vec<String>>, remove: Option<Vec<String>> },
}
```

## Testing Strategy

| 层次                | 测试                        | 框架              |
| ------------------- | --------------------------- | ----------------- |
| Usecase 单元测试    | 参数校验、空输入、超限      | Go `testing`      |
| Repository 集成测试 | 单条 SQL 批量操作正确性     | SQLite `:memory:` |
| Handler E2E         | HTTP 请求→响应完整链路      | Go `httptest`     |
| Web gateway 测试    | request 构建、response 解析 | Vitest            |
| CLI 单元测试        | 参数解析、响应格式化        | Rust `#[test]`    |

批量操作的核心测试点在 Repository 层——确保单条 SQL 的 `IN (...)` 在 SQLite 和 Postgres 下行为一致。

## Boundaries

- **Always do**: 校验 uuid 数量 ≤ 100；owner_user_id 隔离；`IN` 单条 SQL 而非 for 循环
- **Ask first**: schema 变更；添加新依赖；修改响应格式
- **Never do**: 逐条 for 循环提交事务；跳过权限校验；返回明文 uuid 不存在的内部信息

## Success Criteria

1. `POST /memos/batch/archive` 接收 `["u1","u2","u3"]`，单条 SQL 归档且返回 `{ "succeeded": [...], "failed": [...] }`
2. `POST /memos/batch/delete` 行为同上（物理删除）
3. `POST /memos/batch/tag` 接收 `{ uuids, add, remove }`，单条 SQL 完成打标/去标
4. 超过 100 条返回 400
5. 非当前用户 uuid 出现在 `failed` 列表而非静默跳过
6. Web Memo 列表支持 checkbox 多选 + 批量操作栏
7. CLI `daily batch archive --ids u1,u2` 可用
8. 100 条批量操作的 P99 延迟 ≤ 200ms（与单条操作在同一数量级）

## Transaction Model

批量操作的事务流程：

```
1. 预校验：SELECT uuid FROM memo WHERE uuid IN ($uuids) AND owner_user_id = $uid
   → 得到"实际可操作"的 uuid 列表
   → 差异部分进入 failed (not_found / forbidden)
2. 开启事务
3. 执行 SQL：UPDATE / DELETE / INSERT ... WHERE uuid IN ($valid_uuids)
4. 提交事务
5. 返回 { succeeded: valid_uuids, failed: precheck_diffs }
```

全部在单次事务中完成。失败来自预校验阶段，不是执行阶段逐条试错。`skipped`（已归档/已删除/标签已存在）返回在 `succeeded` 中——语义上是"请求了但无需额外操作"，不算错误。

## Future: Shopping Cart 模式

当前方案用 checkbox 多选 + 立即执行。后续可考虑购物车模式：

- Memo 列表每行有"加入批处理"按钮
- 侧边栏或底部有一个浮动"批处理篮"，显示已选数量
- 点击批处理篮展开选择：归档/删除/打标
- 优势：用户跨页面浏览时持续积累选中项，不打断浏览流
- 复杂度：需要跨路由持久化选中状态（Pinia），与现有 `selectedMemo` 逻辑可能冲突

当前不实现。作为设计决策记录，如需评估可新建 spec。

## Decisions

1. `batch/delete` **物理删除**。归档已有单独 endpoint（`DELETE /memos/:uuid`），批量版本同理。
2. Web 多选用 **checkbox 列**。显式可发现，交互成本最低。购物车模式作为 future option。
3. **非幂等但宽容**：已归档/已删除的数据重复提交不报错，归入 `succeeded` 列表。只有"不存在且不属于当前用户"的数据才进入 `failed`。

---

## Plan

### Component Dependency Graph

```
SQL queries (sqlc) ──→ Repository interface/impl ──→ Usecase ──→ Handler ──→ Router
                              │                                                    │
                              │                                          Server API complete
                              │                                                    │
                              ├──────────────────── Web ─────────────── CLI ───────┤
                              │     Gateway → Service → UI               args → handler
                              │                                                    │
                              └──────────────────── Doc ───────────────────────────┘
                                    api/ docs update + doc-route-check
```

Server 完成后，Web 和 CLI 可并行。Docs 在最后。

### Implementation Order

1. **SQL queries**（sqlc `.sql` 文件）— 3 条批量 SQL 写入 SQLite + Postgres 各自的 query 文件
2. **Repository interface**（`port/`）— 新增 `BatchArchive`, `BatchDelete`, `BatchTag`
3. **Repository impl**（persistence/sqlite + postgres）— 调用 sqlc 生成的批量方法
4. **DTO + Usecase** — `dto.BatchResult`, `MemoService.BatchArchive/Delete/Tag`
5. **Handler + Router** — 3 个 handler + 路由注册
6. **Server tests** — usecase 单元 + repository 集成 + handler E2E
7. **Web gateway + service** — `HttpBatchGateway`, `BatchService`
8. **Web UI** — MemoList checkbox + BatchActionBar
9. **CLI** — args + models + trait + api impl + handler
10. **Docs** — api/ 更新 + `make doc-route-check` 验证

### Risks

| 风险                                                      | 概率 | 影响 | 缓解                                                       |
| --------------------------------------------------------- | ---- | ---- | ---------------------------------------------------------- |
| sqlc `IN (sqlc.slice())` 在 SQLite 和 Postgres 语法不一致 | 中   | 高   | 先写 Postgres 测试通过的 SQL，SQLite 端单独适配            |
| `batch/tag` 的 `add` + `remove` 组合需要先删后插          | 低   | 中   | 在事务中分两步执行：先 DELETE 再 INSERT                    |
| Web UI 现有选中逻辑冲突                                   | 低   | 中   | 新增独立 `selectedBatchIds` 状态，不与 `selectedMemo` 耦合 |

## Tasks

### Task 1: Server SQL queries + sqlc

- [ ] **1a**: Write SQLite batch SQL in `persistence/sqlite/query/batch.sql`
  - `--name: BatchArchive` → `UPDATE memo SET row_status = 'archived' ... WHERE uuid IN (sqlc.slice('uuids'))`
  - `--name: BatchDelete` → `DELETE FROM memo WHERE uuid IN (sqlc.slice('uuids'))`
  - `--name: BatchTagAdd` → `INSERT INTO memo_tag ... SELECT ... ON CONFLICT DO NOTHING`
  - `--name: BatchTagRemove` → `DELETE FROM memo_tag WHERE memo_uuid IN (sqlc.slice('uuids')) AND tag_name IN (sqlc.slice('tags'))`
  - **Acceptance**: sqlc generate 成功，生成类型安全 Go 函数
  - **Verify**: `make sqlc-gen` 不报错

- [ ] **1b**: Write Postgres batch SQL in `persistence/postgres/query/batch.sql`
  - 同上，适配 pgx 语法（`= ANY($1)` / `sqlc.slice()`）
  - **Acceptance**: 同 1a
  - **Verify**: `make sqlc-gen` 不报错

### Task 2: Server Repository interface + impl

- [ ] **2a**: `port/` 新增 `BatchArchive(ctx, userID, uuids) ([]string, error)` 等方法
  - 方法返回实际成功操作的 uuid 列表（用于 succeeded）
  - **Acceptance**: 接口定义明确，返回成功列表而非计数
  - **Verify**: 编译通过

- [ ] **2b**: SQLite Repository 实现批量方法（调用 sqlc 生成的 `q.BatchArchive`）
- [ ] **2c**: Postgres Repository 实现批量方法
  - **Acceptance**: 三个方法在两种 DB 下都实现
  - **Verify**: `go build ./...` 通过

### Task 3: Server DTO + Usecase

- [ ] **3a**: `dto/` 新增 `BatchResult { Succeeded, Failed []FailedItem }`、`BatchTagRequest`
- [ ] **3b**: `MemoService` 新增 `BatchArchive/BatchDelete/BatchTag`
  - 校验：empty → 400, >100 → 400
  - 预校验所有权 → 拆分为 succeeded/failed
  - 单事务调用 Repository
  - **Acceptance**: 三个 usecase 完整实现，含参数校验和错误分类
  - **Verify**: `go test ./internal/application/...` 通过

### Task 4: Server Handler + Router

- [ ] **4a**: `memo_handler.go` 新增 `BatchArchive/BatchDelete/BatchTag` handler
  - 解析 JSON body → 调用 usecase → PresentJSON
  - **Acceptance**: handler 正确处理 JSON 输入，返回结构化响应
  - **Verify**: `go build ./...` 通过

- [ ] **4b**: `router.go` 注册三个新路由（挂到 `secured` group）
  - **Acceptance**: 路由注册到正确 group，有 auth middleware
  - **Verify**: `go build ./...` 通过

### Task 5: Server Tests

- [ ] **5a**: Usecase 单元测试 — 空输入、超限、预校验
- [ ] **5b**: Repository 集成测试 — SQLite `:memory:` 下验证 SQL 正确性
- [ ] **5c**: Handler E2E 测试 — `httptest` 完整请求 → 响应
  - **Acceptance**: 三个层次的测试覆盖核心路径和边界
  - **Verify**: `make test-sl` 通过

### Task 6: Web Gateway + Service

- [ ] **6a**: `IBatchPort` + `IBatchStore` 接口定义
- [ ] **6b**: `HttpBatchGateway` 实现三个 POST 调用
- [ ] **6c**: `BatchService` 封装参数校验 + 调用 gateway
  - **Acceptance**: gateway 构建正确 JSON、解析 BatchResponse
  - **Verify**: `pnpm test` 通过

### Task 7: Web MemoList checkbox + BatchActionBar

- [ ] **7a**: MemoList 每行新增 checkbox，`v-model` 绑定到独立 `selectedBatchIds: Set<string>`
  - 与现有选中行为（点击 → 编辑）不冲突：checkbox 区域独立事件
- [ ] **7b**: 底部或顶部浮出 BatchActionBar，含"归档"、"删除"、"打标"按钮
  - 选中 0 条时隐藏，≥1 条时显示
  - 显示已选数量
- [ ] **7c**: BatchActionBar 按钮点击 → 调用 BatchService → 显示结果 toast
  - 成功/失败分别用 TDesign MessagePlugin
  - **Acceptance**: 选择 → 操作 → 反馈完整链路
  - **Verify**: 手动测试：选 3 条归档，确认列表刷新

### Task 8: CLI batch commands

- [ ] **8a**: `args.rs` 新增 `BatchCommands` enum（Archive/Delete/Tag）
- [ ] **8b**: `ports.rs` 新增 `BatchOps` trait（或挂到 `MemoClient`）
- [ ] **8c**: `models.rs` 新增 `BatchRequest` / `BatchResponse`
- [ ] **8d**: `api.rs` 实现三个 POST 调用
- [ ] **8e**: `handlers.rs` 新增 `handle_batch`
- [ ] **8f**: `main.rs` 注册路由
  - **Acceptance**: `daily batch archive --ids u1,u2` 可用
  - **Verify**: `cargo test` + `cargo build` 通过

### Task 9: Docs

- [ ] **9a**: `server/docs/api/memo.md` 新增三个 batch endpoint 文档
- [ ] **9b**: `web/docs/business/CAPABILITIES.md` 新增 batch 消费记录
- [ ] **9c**: `cli/docs/CAPABILITIES.md` 新增 batch 命令记录
  - **Acceptance**: 三个文档都同步
  - **Verify**: `make doc-route-check` + `bash scripts/docroute-check.sh` 通过
