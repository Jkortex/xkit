# ADR-0010 Current Frontend Layering

> 当前前端分层结构：presentation / application/services / application/port / infra

## Status

Accepted

## Context

`packages/daily/web` 之前积累过多份关于事件、projection、read-model 的 ADR，但实际实现已经收敛到了另一条更简单的主路径：

- 写路径不再以 application event 为中心
- 读侧投影细节主要收在 `infra/stores`
- `Service` 通过 `IMemoStore` 等 port 协调本地状态
- `hook` 只订阅 store，通过窄接口触发动作

旧文档（`architecture.md`）中关于 V2.0 以 Service 为中心的描述基本准确，但已被本 ADR 取代。

## Decision

当前前端分层以以下结构为准：

1. `presentation`
2. `application/services`
3. `application/rules`
4. `application/port`
5. `infra`

## Layer Responsibilities

### 1. `presentation`

职责：

- Vue 页面与组件
- 页面装配 hook
- 局部状态 hook
- presenter / view-model
- UI command

输入：

- 用户交互
- `service` 暴露的能力
- `store` 响应式数据

输出：

- Vue `emits`
- UI command
- 对 `service` 的调用

约束：

- 不定义业务规则
- 不直接依赖 gateway 实现细节
- hook 不直接持有完整 store 写接口

### 2. `application/services`

职责：

- 编排业务流程
- 调用 gateway/port
- 调用 `application/rules` 做业务校验
- 成功后将结果写入 store（write-through cache）

输入：

- 来自 presentation 的动作参数
- `port`
- store 接口（作为缓存层写入）

输出：

- DTO / Result
- 写入 store 更新客户端状态

约束：

- 不依赖 Vue / DOM
- 不维护长期读模型状态

### 3. `application/rules`

职责：

- 纯业务校验函数
- 可被 service 和 UI 层共同复用

典型例子：

- `validateMemoContent`
- 标签合并约束

约束：

- 无副作用纯函数

### 4. `application/port`

职责：

- 定义 application 依赖的抽象接口与稳定 DTO

包含：

- gateway contract（如 `IMemoPort`）
- store contract（如 `IMemoStore`）
- DTO 类型定义

原则：

- hook 使用最小接口
- 不把 store 实现细节暴露给上层

### 5. `infra`

职责：

- HTTP/gateway 实现
- store 实现（Pinia）
- container / 依赖装配

当前约定：

- `infra/stores` 是客户端状态 owner
- `memo / tag / stats / auth` 的本地状态收在这里
- store 对 application 层暴露为 `IMemoStore` 等 port 接口

## Main Flows

### Write Path

`presentation -> service -> gateway -> store (write-through)`

含义：

- 写流程在 service 中编排
- gateway 调用成功后，结果直接写入 store 缓存
- 调用方可从 store 读取最新状态

### Read Path

`presentation -> service -> gateway -> store (read-through)`

或：

`presentation -> store (直接响应式绑定)`

含义：

- service 负责首次加载或回源刷新
- store 持有当前快照，UI 直接响应式绑定

## Naming

当前命名约定：

- 编排层：`XxxService`
- 规则：`xxx-rules`
- store 接口：`IXxxStore`
- gateway 接口：`IXxxPort`
- store 实现：`useXxxStore`
- gateway 实现：`HttpXxxGateway`
- 页面桥接：`useXxxModel`
- 页面装配：`useXxxView`
- 局部状态：`useXxxState`
- 副作用/快捷键：`useXxxEffects`
- 业务动作：`useXxxActions`

## Consequences

正向影响：

- 主路径更直接，便于理解和落位
- store 作为 write-through cache 层，简化数据流
- 不再需要维护一整套未落地的 event/projection 设计

代价：

- `store` 承担缓存和响应式状态双重职责，需注意边界
- 需要通过测试和文档守住边界，避免 hook 直接拿 store 写接口

## Follow-up

1. 以本 ADR 为准，删除已经不再代表当前实现的旧 ADR 和文档。
2. 后续新增 service 时，优先沿用 write-through cache 模式：
   - Service 编排业务流程
   - Store 作为共享缓存层
   - Presentation 直接响应式绑定 store
3. 继续通过边界测试约束层间依赖。
