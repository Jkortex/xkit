# Daily Web 接口调用与业务能力地图

本文档记录了 Daily Web 当前消费的后端 API 及其业务语义。用于与 Daily Server 对齐，并在接口契约发生变更时作为检查基准。

说明：

- 基线路由前缀默认为 `/api/v1`（详见 `infra/http/FetchClient.ts`）。
- 接口调用通过 `application/ports` 定义契约，在 `infra/gateway` 中实现。

---

## 1. 认证与访问控制

Web 端主要基于 Cookie Session 进行认证，部分场景涉及邀请制流程。

| 业务能力     | API 调用                         | 关键参数                       | 实现位置                     | 语义与备注                                 |
| :----------- | :------------------------------- | :----------------------------- | :--------------------------- | :----------------------------------------- |
| 获取当前用户 | `GET /auth/me`                   | -                              | `AuthStore.bootstrap`        | 应用初始化时调用，用于判定登录态及角色。   |
| 登录         | `POST /auth/login`               | `username`, `password`         | `AuthStore.login`            | 成功后由服务端写入 `sid` Cookie。          |
| 退出登录     | `POST /auth/logout`              | -                              | `AuthStore.logout`           | 清理 Session，同时重置本地 Store 状态。    |
| 校验邀请码   | `GET /auth/invites/:code/verify` | `code` (path)                  | `AuthStore.verifyInvite`     | 注册页预校验，返回有效性、角色及过期时间。 |
| 邀请注册     | `POST /auth/register-by-invite`  | `code`, `username`, `password` | `AuthStore.registerByInvite` | 使用邀请码直接注册并建立登录态。           |

### 1.1 邀请码管理（管理员权限）

| 业务能力       | API 调用                        | 关键参数            | 实现位置                     | 语义与备注                         |
| :------------- | :------------------------------ | :------------------ | :--------------------------- | :--------------------------------- |
| 查询邀请码列表 | `GET /auth/invites`             | `status`, `limit`   | `AuthStore.listInvites`      | 支持按状态过滤，默认 `limit=100`。 |
| 创建邀请码     | `POST /auth/invites`            | `role`, `ttl_hours` | `AuthStore.createInvite`     | 创建成功后响应包含明文 `code`。    |
| 撤销邀请码     | `POST /auth/invites/:id/revoke` | `id` (path)         | `AuthStore.revokeInvite`     | 仅活跃状态邀请码可撤销。           |
| 查询配额摘要   | `GET /auth/invites/summary`     | -                   | `AuthStore.getInviteSummary` | 返回当前活跃用户数与配额上限。     |

### 1.2 API Key 管理

| 业务能力     | API 调用                     | 关键参数                        | 实现位置                      | 语义与备注                                   |
| :----------- | :--------------------------- | :------------------------------ | :---------------------------- | :------------------------------------------- |
| 创建 API Key | `POST /auth/api-keys`        | `label`, `ttl_hours`            | `ApiKeyStore.createKey`       | 为当前用户创建 API Key，响应包含一次性明文。 |
| 创建 (直接)  | `POST /auth/api-keys/direct` | `username`, `password`, `label` | `ApiKeyStore.createKeyDirect` | 使用凭据直接换取 API Key，无需先行登录。     |
| 查询列表     | `GET /auth/api-keys`         | -                               | `ApiKeyStore.fetchKeys`       | 返回当前用户所有 API Key 元数据。            |
| 删除 Key     | `DELETE /auth/api-keys/:id`  | `id` (path)                     | `ApiKeyStore.deleteKey`       | 撤销指定 API Key。                           |
| 自注销       | `POST /auth/api-keys/revoke` | -                               | `ApiKeyStore.revokeCurrent`   | 撤销当前请求所用的 API Key。                 |

_注：Web 端主要使用 Cookie Session 认证；API Key 主要用于 CLI 或第三方集成。_

---

## 2. Memo 核心业务

### 2.1 采集与生命周期

| 业务能力  | API 调用              | 关键参数                         | 实现位置                        | 语义与备注                                                                                          |
| :-------- | :-------------------- | :------------------------------- | :------------------------------ | :-------------------------------------------------------------------------------------------------- |
| 创建 Memo | `POST /memos`         | `content`, `resource_ids`, `ttl` | `HttpMemoGateway.createMemo`    | `content` 尾部嵌入 `#tag` 标签块，服务端自动提取并分离存储；支持绑定已上传的资源 ID，支持可选 TTL。 |
| 获取列表  | `GET /memos`          | 见下文                           | `HttpMemoGateway.getMemos`      | 分页获取当前用户正常状态的 Memo。返回 `content` 已清洗（不含尾部标签），`tags` 为独立数组。         |
| 更新 Memo | `PATCH /memos/:uuid`  | `content`, `resource_ids`, `ttl` | `HttpMemoGateway.updateMemo`    | 格式同创建：`content` 尾部嵌入标签，服务端重新提取、更新存储。                                      |
| 删除 Memo | `DELETE /memos/:uuid` | -                                | `HttpMemoGateway.deleteMemo`    | 软删除（归档）操作。                                                                                |
| 随机漫步  | `GET /memos/random`   | -                                | `HttpMemoGateway.getRandomMemo` | 随机获取一条笔记，返回同列表格式。                                                                  |

### 2.2 标签块格式

服务端使用**严格模式**提取标签：

- **标签块位置**：位于 content 末尾，由连续的以 `#` 开头的行组成
- **分隔规则**：正文与标签块之间不需要空行，服务端从底部向上扫描
- **校验规则**：每行所有空格分隔项必须是合法标签（`#work`），空标签 `#` 或含非法符号返回 400
- **长度限制**：标签名 ≤ 32 字符
- **内容清洗**：存储时自动从 `content` 中剔除尾部标签块，`content` 正文中的 `#inline` 视为普通文本不提取
- **别名展开**：服务端在存储前将别名标签解析为规范化标签

前端发送示例：

```
今天完成了架构重构
#work #ops
```

存储后查询返回：

```json
{
  "content": "今天完成了架构重构",
  "tags": ["work", "ops"]
}
```

编辑时：前端需将 `tags` 拼回 `content` 尾部再发送。

### 2.3 检索参数映射 (`MemoListQuery`)

`GET /memos` 调用时支持以下映射：

| Web 参数           | 后端 API 参数       | 说明                                                       |
| :----------------- | :------------------ | :--------------------------------------------------------- |
| `search`           | `search`            | 全文检索字符串。                                           |
| `tag`              | `tag`               | 单标签过滤。                                               |
| `from`             | `from`              | 格式 `YYYY-MM-DD`。                                        |
| `to`               | `to`                | 格式 `YYYY-MM-DD`。                                        |
| `hasResource`      | `has_resource`      | `bool` 字符串。                                            |
| `includeResources` | `include_resources` | 是否在响应中附带资源详情。                                 |
| `tagsAny`          | `tags_any`          | 逗号分隔的标签。                                           |
| `tagsAll`          | `tags_all`          | 逗号分隔的标签。                                           |
| `tagsExclude`      | `tags_exclude`      | 逗号分隔的排除标签（NOT）。                                |
| `sort`             | `sort`              | `created_at_desc` / `created_at_asc` / `updated_at_desc`。 |
| `limit`            | `limit`             | 分页大小。                                                 |
| `offset`           | `offset`            | 偏移量。                                                   |

### 2.4 历史与回滚

| 业务能力     | API 调用                          | 关键参数             | 实现位置                          | 语义与备注                                                         |
| :----------- | :-------------------------------- | :------------------- | :-------------------------------- | :----------------------------------------------------------------- |
| 获取历史列表 | `GET /memos/:uuid/history`        | `uuid` (path)        | `HttpMemoGateway.listMemoHistory` | 获取该 Memo 的历史快照列表。历史快照保留旧 content（含当时标签）。 |
| 执行回滚     | `POST /memos/:uuid/rollback/:hid` | `uuid`, `hid` (path) | `HttpMemoGateway.rollbackMemo`    | 将 Memo 回滚至指定历史版本，回滚后 content/tag 重新分离存储。      |

### 2.5 临时 Memo / TTL

| 场景             | 语义                                                         |
| :--------------- | :----------------------------------------------------------- |
| 显式 TTL         | `POST/PATCH /memos` 传 `ttl`，例如 `1h`、`3d`。              |
| `#temp` 兜底 TTL | 若标签块包含 `#temp` 且未显式传 `ttl`，默认 3 天过期。       |
| 过期处理         | 后台任务自动归档已过期内容。                                 |
| 检索特性         | 带有 TTL 的 Memo 被视为临时内容，不参与全文检索 (`search`)。 |

---

## 3. 标签治理

| 业务能力        | API 调用                      | 关键参数             | 实现位置                         | 语义与备注                   |
| :-------------- | :---------------------------- | :------------------- | :------------------------------- | :--------------------------- |
| 标签统计列表    | `GET /tags`                   | -                    | `HttpMemoGateway.getTags`        | 返回所有标签及其使用计数。   |
| 标签重命名/合并 | `POST /tags/rename`           | `from`, `to`         | `HttpMemoGateway.renameTag`      | 若 `to` 已存在则执行合并。   |
| 标签批量合并    | `POST /tags/merge`            | `sources`, `target`  | `HttpMemoGateway.mergeTags`      | 将多个源标签合并至目标标签。 |
| 创建/更新别名   | `POST /tags/aliases`          | `alias`, `canonical` | `HttpMemoGateway.upsertTagAlias` | 建立输入规范化规则。         |
| 获取别名列表    | `GET /tags/aliases`           | -                    | `HttpMemoGateway.listTagAliases` | 获取当前的别名映射表。       |
| 删除别名        | `DELETE /tags/aliases/:alias` | `alias` (path)       | `HttpMemoGateway.deleteTagAlias` | 移除别名规则。               |
| 获取治理审计    | `GET /tags/audits`            | `limit`, `action`    | `HttpMemoGateway.listTagAudits`  | 查看最近的治理动作日志。     |

### 3.1 当前实现语义

- **标签统计**：按当前用户作用域返回。
- **别名与审计**：当前实现为系统级共享（不带 `user_id`），别名规则对所有用户生效。
- **重命名即合并**：若目标标签已存在，重命名操作会自动触发合并，并返回受影响的 Memo 数。

---

## 4. 标签集 (Tag Sets)

预设筛选条件组合，按分组组织。支持 `tags_any`（OR）、`tags_all`（AND）、`tags_exclude`（NOT）。
启动时自动为管理员创建 3 组默认标签集（工作流 / 阅读管理 / 项目追踪）。

### 4.1 侧栏快捷面板

| 业务能力     | 实现位置               | 语义与备注                                                                        |
| :----------- | :--------------------- | :-------------------------------------------------------------------------------- |
| 展开快捷面板 | `TagSetQuickPanel.vue` | 侧栏底部快速访问入口，展示分组 + 未分组标签集。点击标签集自动跳转首页并应用筛选。 |
| 按分组浏览   | `TagSetQuickPanel.vue` | 展开后按 `group_id` 分组折叠展示。                                                |
| 搜索标签集   | `TagSetQuickPanel.vue` | 面板内实时过滤。                                                                  |

### 4.2 管理页面

| 业务能力    | 实现位置                                               | 语义与备注                     |
| :---------- | :----------------------------------------------------- | :----------------------------- |
| 查看管理页  | `/tag-sets` 路由，`TagSetManageView.vue`               | 全功能管理界面。               |
| 分组 CRUD   | `TagSetService.createGroup/updateGroup/deleteGroup`    | 左侧面板管理分组。             |
| 标签集 CRUD | `TagSetService.createTagSet/updateTagSet/deleteTagSet` | 右侧按组展示、新建/编辑/删除。 |

### 4.3 后端 API 映射

| 业务能力   | API 调用                       | 关键参数                     | 实现位置                         |
| :--------- | :----------------------------- | :--------------------------- | :------------------------------- |
| 列出分组   | `GET /tag-set-groups`          | -                            | `HttpTagSetGateway.listGroups`   |
| 创建分组   | `POST /tag-set-groups`         | `name`, `weight`             | `HttpTagSetGateway.createGroup`  |
| 更新分组   | `PATCH /tag-set-groups/:id`    | `name`, `weight`             | `HttpTagSetGateway.updateGroup`  |
| 删除分组   | `DELETE /tag-set-groups/:id`   | -                            | `HttpTagSetGateway.deleteGroup`  |
| 列出标签集 | `GET /tag-sets?group_id=`      | `group_id`                   | `HttpTagSetGateway.listTagSets`  |
| 创建标签集 | `POST /tag-sets`               | `name`, `group_id`, `tags_*` | `HttpTagSetGateway.createTagSet` |
| 更新标签集 | `PATCH /tag-sets/:id`          | `name`, `tags_*`             | `HttpTagSetGateway.updateTagSet` |
| 删除标签集 | `DELETE /tag-sets/:id`         | -                            | `HttpTagSetGateway.deleteTagSet` |
| 应用标签集 | `TagSetQuickPanel.applyTagSet` | -                            | 前端拼接筛选参数跳转 `/` 路由。  |

---

## 5. 资源与系统管理

| 业务能力 | API 调用              | 关键参数           | 实现位置                         | 语义与备注                           |
| :------- | :-------------------- | :----------------- | :------------------------------- | :----------------------------------- |
| 上传资源 | `POST /resources`     | `file` (multipart) | `HttpResourceGateway.upload`     | 返回资源元数据，含 ID。              |
| 访问资源 | `GET /resources/:id`  | `id` (path)        | `ResourcePresenter.toViewModel`  | 前端拼接 URL，由浏览器直接发起请求。 |
| 导出数据 | `GET /system/export`  | -                  | `HttpResourceGateway.exportData` | 获取完整备份 ZIP。                   |
| 导入数据 | `POST /system/import` | `file` (multipart) | `HttpResourceGateway.importData` | 上传备份 ZIP，返回结构化报告。       |

---

## 6. 统计与回顾

| 业务能力     | API 调用     | 关键参数 | 实现位置                    | 语义与备注                 |
| :----------- | :----------- | :------- | :-------------------------- | :------------------------- |
| 获取全量统计 | `GET /stats` | -        | `HttpStatsGateway.getStats` | 返回总量计数与热力图数据。 |

---

## 7. 数据结构对齐 (DTO)

### 7.1 Memo 对象 (`BackendMemoDTO`)

```typescript
{
  uuid: string;
  content: string;        // 已清洗正文——不含尾部 #tag，正文内的 #inline 保留
  row_status: string;     // 'normal' | 'archived'
  tags: string[] | null;  // 独立标签数组，由服务端从 content 尾部提取
  resources: {
    id: string;
    filename: string;
    size: number;
    mime_type: string;
    created_at: string;
  }[] | null;
  expires_at: string | null;
  headline?: string;
  created_at: string;
  updated_at: string;
}
```

前端 `transformMemo` 处理：

- `content` 透传（已是清洗后内容）
- `tags` 空值归一化为 `[]`（`raw.tags || []`）

### 7.2 导入报告 (`BackendImportReportDTO`)

```typescript
{
  message: string;
  memos_imported: number;
  resources_imported: number;
  memos_skipped: number;
  resources_skipped: number;
  report: {
    memos: { imported: number; skipped: number; details: any[] };
    resources: { imported: number; skipped: number; details: any[] };
  };
}
```

### 7.3 统计数据 (`BackendStatsDTO`)

```typescript
{
  memos_total: number;
  tags_total: number;
  resources_total: number;
  heatmap: {
    date: string;
    count: number;
  }
  [] | null;
}
```
