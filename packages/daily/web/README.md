# Daily Web (V1.0)

基于 Vue 3 + Tailwind v4 + TDesign 的极简个人笔记门户。

## 🛠️ 技术栈

- **框架**: Vue 3.5 (Composition API)
- **样式**: TailwindCSS v4 + TDesign Vue Next
- **状态**: Pinia
- **工具**: oxlint, oxfmt (Oxc 工具链)
- **架构**: 严格 Clean Architecture (Gateway, Transform, Rules, UseCase, Presenter)

## 🚀 快速开始

### 1. 安装依赖

```bash
pnpm install
```

### 2. 开发环境运行

确保后端已在 `8080` 端口启动：

```bash
pnpm dev
```

### 3. 代码质量

```bash
pnpm lint  # 极速 Lint
pnpm fmt   # 格式化
pnpm test  # 运行 Vitest 单元测试
```

## 📂 架构说明

详见：

- `docs/architecture.md`
- `docs/ui_guidelines.md`
- `docs/hotkey_modes.md`
