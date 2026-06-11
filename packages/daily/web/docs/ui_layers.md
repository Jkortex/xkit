# Daily Web UI Layer 规范

## 1. 目的

统一页面层级语义，避免组件各自定义 `z-index` 导致遮挡、穿透和弹窗叠层冲突。

## 2. Token 与层级语义

只允许使用以下 token：

- `--daily-z-sticky`: 吸顶/固定区（如顶部工具条）
- `--daily-z-fab`: 浮动操作按钮
- `--daily-z-overlay`: 遮罩层
- `--daily-z-dialog`: 对话框主体

禁止在组件中写任意数字 `z-index`（含 Tailwind 任意值）。

## 3. 组件约束

- 弹窗类组件默认 `attach="body"`，避免受局部 stacking context 影响。
- Home 主容器不承担全局层级管理职责，避免给根节点设置高 `z-index`。
- 遮罩和弹窗必须成对出现，且 `dialog > overlay > page content`。

## 4. 实施建议

- 复用频次 >= 3 的层级样式，抽到 `@layer components`。
- 在 PR 检查项中加入“是否新增硬编码 z-index”。
- 如需新增层级 token，先在规范文档补充语义，再落地到 `style.css`。
