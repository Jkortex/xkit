# Daily Web UI 规范 (V1.0)

## 1. 目标

统一 Daily 前端的视觉语言与实现方式，保证：

- 明暗主题行为一致。
- 新页面接入成本低。
- 样式修改可控，避免局部“回旧风格”。

配套文档：

- 层级规范：`web/docs/ui_layers.md`

---

## 2. 主题系统规范

### 2.1 架构约束

- 主题由 `data-theme` 驱动，不在组件里写硬编码明暗分支。
- 主题切换统一走 `themeManager`，不在组件里直接读写 `localStorage`。
- 组件只消费语义变量（`var(--...)`），不直接写具体色值。

### 2.2 当前主题

- `calm-light`
- `calm-dark`

主题定义与切换入口：

- `src/presentation/theme/themeManager.ts`
- `src/style.css`

### 2.3 新增主题流程

1. 在 `themeManager.ts` 的 `THEME_PRESETS` 增加主题 id。
2. 在 `style.css` 增加 `:root[data-theme='new-theme']` token 值。
3. 不改业务组件样式类。

---

## 3. Token 约定

### 3.1 基础语义变量

- 品牌：`--td-brand-color`、`--td-brand-color-light`
- 背景：`--td-bg-color-page`、`--td-bg-color-container`、`--td-bg-color-container-hover`
- 边框：`--td-border-level-1-color`、`--td-border-level-2-color`
- 文本：`--td-text-color-primary`、`--td-text-color-secondary`、`--td-text-color-placeholder`

### 3.2 Daily 扩展变量

- Logo：`--daily-logo-bg`、`--daily-logo-fg`
- 页面氛围：`--daily-body-gradient-a`、`--daily-body-gradient-b`
- 热力图：`--daily-heat-0..3`

---

## 4. 组件样式层约定 (`@layer components`)

统一在 `src/style.css` 的 `@layer components` 维护复用类。

### 4.1 前缀与命名

- 前缀统一 `ui-`
- 命名语义优先，不以页面命名

### 4.2 当前核心复用类

- 导航/侧栏：`ui-nav-link`、`ui-sidebar-action`
- 弹窗：`ui-dialog-*`（`body/caption/section/title/label/actions/select`）
- 列表容器：`ui-list-shell`、`ui-list-row`、`ui-list-empty`
- 命令区：`ui-command-*`、`ui-search-shell`、`ui-shortcut-board`
- 尺寸基线：`ui-input-md`、`ui-btn-icon-sm`、`ui-btn-chip-sm`、`ui-btn-fab-lg`
- 卡片与状态：`ui-surface-card`、`ui-surface-card-hover`、`ui-stat-card`、`ui-loading-card`、`ui-empty-board`

---

## 5. 尺寸基线

### 5.1 输入与按钮

- 输入中尺寸：`ui-input-md`
- 图标小按钮：`ui-btn-icon-sm`
- 胶囊筛选按钮：`ui-btn-chip-sm`
- 移动端浮动主按钮：`ui-btn-fab-lg`

### 5.2 间距和圆角（默认）

- 间距基线：`4 / 8 / 12 / 16 / 24 / 32`
- 常用圆角：`0.45rem / 0.65rem / 0.8rem / 1rem`

---

## 6. 代码规则

- 重复 3 次以上的样式组合，必须抽到 `@layer components`。
- 组件内 `scoped` 样式只保留真正局部、不可复用样式。
- 禁止新增 `dark:` 分支；一律改为 token 方案。
- 禁止在组件中写与主题相关的硬编码色值。

---

## 7. 视觉回归检查清单

每次样式改动后，按清单手工检查：

### 7.1 主题一致性

- `calm-light` / `calm-dark` 下页面结构和信息层级一致。
- 文本、边框、背景对比可读（尤其 placeholder、caption）。
- 主题切换后热力图、Logo、卡片状态正确。

### 7.2 交互状态

- 按钮的 `hover/active/disabled/loading` 行为统一。
- 输入框 `focus` 边框与外轮廓一致。
- 弹窗滚动区域在窄窗口下可用。

### 7.3 页面覆盖

- Home: 搜索区、列表卡片、空态、加载态。
- Sidebar: 导航、主题按钮、导入导出入口。
- Stats: 热力图容器、统计卡片。
- Dialog: 导入报告、标签治理。

### 7.4 响应式

- 移动端（<=768）浮动按钮可点击区域充足。
- 侧栏隐藏后主内容无异常留白。
- 弹窗在小屏无横向溢出。
