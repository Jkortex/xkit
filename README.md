# xkit

> 📓 个人知识管理（PKM）应用 monorepo | Vue 3 + Go 全栈

Daily 是一个专注于个人碎片化知识管理的本地化（local-first）工具。采用 Clean Architecture 架构，基于 SQLite 提供轻量、零依赖的本地存储体验。

## 🛠️ 技术栈

| 层面 | 技术 |
|---|---|
| **前端** | Vue 3.5 (Composition API), TypeScript 6, Pinia, Vue Router 5, TDesign Vue Next, TailwindCSS 4 |
| **后端** | Go 1.24+, Gin, SQLite (FTS5), sqlc, Bubble Tea TUI, Cobra CLI |
| **代码质量** | oxlint, oxfmt (Oxc 工具链) |
| **构建** | Vite 8, pnpm 11, rolldown |
| **测试** | Vitest 4, happy-dom, vue-tsc 3 |

## 📦 包结构

| 目录 | 说明 |
|---|---|
| `packages/daily/web/` | Vue 3 SPA 前端 |
| `packages/daily/server/` | Go 后端服务 (Gin, SQLite, TUI, CLI) |
| `packages/hotkeys/` | `@xkit/hotkeys` — Vue 3 快捷键绑定与命令面板库 |
| `packages/strata/` | `@xkit/strata` — Go 分层配置库 (环境变量/文件) |
| `packages/skills/` | 20+ 个 agent skill 包 |
| `packages/pi-kit/` | `@xkit/pi-kit` — pi-coding-agent 扩展 |
| `packages/pi-theme-everforest/` | pi-coding-agent 的 Everforest 护眼主题 |

## 🚀 快速开始

### 前置要求

- Node.js >= 20
- pnpm 11 (`npm install -g pnpm@11`)
- Go 1.24+

### 安装

```bash
pnpm install
```

### 开发

```bash
# 启动前端开发服务器（需要后端在 8080 端口运行）
pnpm dev:web

# 或通过 pnpm xkit 启动完整的开发环境
pnpm xkit dev
```

### 代码质量

```bash
pnpm lint       # oxlint 极速 Lint
pnpm fmt        # 检查格式化
pnpm fmt:fix    # 自动格式化
pnpm typecheck  # TypeScript 类型检查
pnpm ci         # 完整本地 CI（typecheck → test → vet → lint → fmt）
```

### 构建

```bash
pnpm build:web       # 构建前端
pnpm build:hotkeys   # 构建 @xkit/hotkeys
pnpm xkit build      # 构建 Go 后端
```

### 测试

```bash
pnpm test:web        # 前端测试 (Vitest)
pnpm test:hotkeys    # hotkeys 库测试 (Vitest)
pnpm test:cli        # Go CLI 测试
pnpm xkit test       # Go 后端测试
```

## 🏗️ 架构

前后端均采用严格的 **Clean Architecture** 分层：

| 层级 | web (Vue) | server (Go) |
|---|---|---|
| 最外层 | `presentation/` (组件、视图) | `interfaces/` (handler、CLI、TUI) |
| 中间层 | `application/` (用例) | `application/` (用例) |
| 中间层 | `domain/` (实体) | `domain/` (实体) |
| 最内层 | `infra/` (网关、存储) | `infrastructure/` (持久化) |

- 前端使用 `Result<T>` monad（`Success` / `Failure`）替代异常抛出
- `@xkit/hotkeys` 使用静态注册 + 动态上下文节点，快捷键匹配优先深度节点
- `@xkit/strata` 提供分层配置加载：struct defaults → `.env` → `~/.xkit/config.json` → 配置文件 → 环境变量

## 📋 常用命令

| 命令 | 说明 |
|---|---|
| `pnpm lint` | oxlint 检查 |
| `pnpm fmt` | oxfmt 格式化检查 |
| `pnpm fmt:fix` | 自动修复格式 |
| `pnpm typecheck` | vue-tsc 类型检查 |
| `pnpm dev:web` | 启动前端开发服务器 |
| `pnpm build:web` | 构建前端 |
| `pnpm build:hotkeys` | 构建 hotkeys 库 |
| `pnpm test:web` | 前端测试 |
| `pnpm test:hotkeys` | hotkeys 测试 |
| `pnpm test:cli` | Go CLI 测试 |
| `pnpm xkit <task>` | Go 服务管理（dev/test/build/vet/sqlc-gen 等） |
| `pnpm ci` | 完整 CI 流水线 |
| `pnpm web-check` | API 端点同步检查 |

## 🔄 CI 流水线

```
pnpm typecheck → pnpm test:hotkeys → pnpm xkit vet → pnpm xkit test → pnpm test:cli → pnpm lint → pnpm fmt
```

- GitHub Actions 中配置了对 `packages/daily/web/` 变更的自动检查
- 推送前运行 `pnpm ci`

## 🗄️ 数据库

- **SQLite** + FTS5 全文搜索
- sqlc 代码生成：`migrations/sqlite/` → `internal/infrastructure/persistence/sqlite/queries/` → `db/`
- 修改查询后运行 `pnpm xkit sqlc-gen`

## 📖 相关文档

- [AGENTS.md](./AGENTS.md) — Agent 视角的项目概览
- [Daily Web 文档](./packages/daily/web/docs/architecture.md)
- [Daily 后端架构](./packages/daily/server/docs/overview.md)
- [Daily API 参考](./packages/daily/server/docs/api.md)
- [@xkit/hotkeys 设计文档](./packages/hotkeys/docs/design.md)

## 📄 许可

[MIT](./LICENSE)
