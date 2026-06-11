# Memo 管理

## 创建

### POST /memos

创建 Memo。自动从 content 尾部提取正式标签块并清洗正文。若同时传 `tags` 字段，则优先使用。

**Request**

| 参数           | 类型     | 必需 | 默认 | 说明                           |
| -------------- | -------- | ---- | ---- | ------------------------------ |
| `content`      | string   | 是   | -    | 正文，尾部可附带标签块         |
| `tags`         | string[] | 否   | -    | 独立标签，优先级高于尾部提取   |
| `resource_ids` | string[] | 否   | `[]` | 已上传资源的 ID 列表           |
| `ttl`          | string   | 否   | null | 过期时间，如 `1h`、`24h`、`3d` |

```json
{
  "content": "Meeting notes\n#Work #Ops",
  "resource_ids": ["sha256-hash-1"],
  "ttl": "3d"
}
```

**Response 201**: MemoResponse。

**Response 400**: 标签格式非法。

---

## 列表/检索

### GET /memos

分页获取当前用户正常状态 Memo。支持多维筛选和全文检索。

**Request**

```
GET /api/v1/memos?search=&tag=&from=&to=&tags_any=&tags_all=&tags_exclude=&has_resource=&include_resources=&sort=&limit=&offset=
```

| 参数                | 类型   | 必需 | 默认              | 说明                                                                                                      |
| ------------------- | ------ | ---- | ----------------- | --------------------------------------------------------------------------------------------------------- |
| `search`            | string | 否   | -                 | 全文检索。中文分词 AND。标签 weight A > 文件名 weight B > 正文 weight D。返回 `headline`（`<mark>` 高亮） |
| `tag`               | string | 否   | -                 | 单标签精确匹配                                                                                            |
| `tags_any`          | string | 否   | -                 | 逗号分隔，任一命中（OR）                                                                                  |
| `tags_all`          | string | 否   | -                 | 逗号分隔，全部命中（AND）                                                                                 |
| `tags_exclude`      | string | 否   | -                 | 逗号分隔，排除命中（NOT）                                                                                 |
| `from`              | string | 否   | -                 | 创建时间起始 `YYYY-MM-DD`                                                                                 |
| `to`                | string | 否   | -                 | 创建时间结束 `YYYY-MM-DD`                                                                                 |
| `has_resource`      | bool   | 否   | -                 | 是否含附件                                                                                                |
| `include_resources` | bool   | 否   | false             | 响应中附带资源详情                                                                                        |
| `sort`              | enum   | 否   | `created_at_desc` | `created_at_desc` / `created_at_asc` / `updated_at_desc`                                                  |
| `limit`             | int    | 否   | 20                | 分页大小，范围 1-100                                                                                      |
| `offset`            | int    | 否   | 0                 | 偏移量                                                                                                    |

`tag`、`tags_any`、`tags_all`、`tags_exclude` 可叠加。`from > to` 返回 400。

**Response 200**: MemoResponse[]。未传 `include_resources` 时 `resources` 为 null。

---

## 获取单条

### GET /memos/:uuid

获取指定 Memo 的详细信息。

**Request**

```
GET /api/v1/memos/:uuid
```

| 参数   | 位置 | 说明      |
| ------ | ---- | --------- |
| `uuid` | path | Memo UUID |

**Response 200**: MemoResponse。

**Response 404**: 不存在。

---

## 状态迁移（Plan 任务）

### POST /memos/:uuid/transition

变更 Memo 的任务状态。标签转换规则：`#todo` → `#doing` → `#done` / `#failed`。执行 `#doing` 时自动追加 `#by/<agent_id>`。

**Request**

| 参数       | 类型   | 必需 | 说明                                  |
| ---------- | ------ | ---- | ------------------------------------- |
| `status`   | string | 是   | 目标状态：`doing` / `done` / `failed` |
| `agent_id` | string | 否   | 执行者标识（转 `doing` 时使用）       |

```json
{ "status": "doing", "agent_id": "agent-001" }
```

**Response 200**: 转换后的 MemoResponse。

**Response 400**: 当前状态不允许转换到目标状态。

**Response 404**: Memo 不存在。

---

## 随机

### GET /memos/random

随机返回一条当前用户 Memo。

**Request**

```
GET /api/v1/memos/random
```

**Response 200**: MemoResponse。

**Response 404**: 无 Memo。

---

## 更新

### PATCH /memos/:uuid

更新 Memo。更新前自动保存当前版本至历史快照。重新提取标签、更新资源绑定。

**Request**

| 参数           | 类型     | 必需 | 说明                     |
| -------------- | -------- | ---- | ------------------------ |
| `content`      | string   | 否   | 新正文                   |
| `tags`         | string[] | 否   | 新标签列表               |
| `resource_ids` | string[] | 否   | 新资源 ID（全量替换）    |
| `ttl`          | string   | 否   | 新 TTL，传 null 清除过期 |

```json
{
  "content": "Updated meeting notes\n#Work",
  "resource_ids": []
}
```

**Response 200**: MemoResponse。

**Response 404**: Memo 不存在。

---

## 归档

### DELETE /memos/:uuid

软删除（`row_status` → `archived`）。

**Request**

```
DELETE /api/v1/memos/:uuid
```

**Response 204**: 无 body。

**Response 404**: 不存在。

---

## 历史

### GET /memos/:uuid/history

获取 Memo 历史快照列表（最近 20 条真实变更）。

**Request**

```
GET /api/v1/memos/:uuid/history
```

**Response 200**

```json
[
  {
    "id": "019e1507-aaaa-7d17-adc8-0785aa7e898f",
    "memo_uuid": "019e1506-e5ea-7d17-adc8-0785aa7e898f",
    "content": "Old version #Work",
    "tags": ["Work"],
    "resource_ids": ["res-1"],
    "created_at": "2026-03-01T14:50:00Z"
  }
]
```

---

## 回滚

### POST /memos/:uuid/rollback/:hid

回滚至指定历史版本。回滚后 content/tag 重新分离存储。

**Request**

```
POST /api/v1/memos/:uuid/rollback/:hid
```

| 参数   | 位置 | 说明          |
| ------ | ---- | ------------- |
| `uuid` | path | Memo UUID     |
| `hid`  | path | 历史版本 UUID |

**Response 200**: MemoResponse。

**Response 404**: uuid 或 hid 不存在。

---

## 标签提取规则

服务端使用**严格模式**：

1. **标签块位置**：content 末尾，连续以 `#` 开头的行
2. **扫描方向**：自底向上
3. **校验**：行内所有空格分隔项必须是合法标签（`#Tag`），否则 400
4. **非法**：空标签 `#`、>32 字符、含非法符号
5. **清洗**：存储时自动剔除尾部标签块及其前置换行
6. **正文保护**：正文中的 `#inline` 视为纯文本，不提取
7. **别名展开**：存储前解析别名 → canonical
8. **容忍空行**：标签块后可存在空行

**发送示例：**

```
今天完成了架构重构
#work #ops
```

**存储后返回：**

```json
{
  "content": "今天完成了架构重构",
  "tags": ["work", "ops"]
}
```

**编辑时**：前端需将 `tags` 拼回 `content` 尾部再发送。

---

## TTL / 临时 Memo

| 场景         | 语义                                   |
| ------------ | -------------------------------------- |
| 显式 TTL     | `POST/PATCH` 传 `ttl`，如 `1h`、`3d`   |
| `#temp` 兜底 | 标签块含 `#temp` 且未传 `ttl`，默认 3d |
| 过期处理     | 后台任务按小时扫描，自动归档           |
| 检索降噪     | `search_vector` 置 NULL，不参与 FTS    |
| `expires_at` | ISO8601 时间戳，null 表示不过期        |

---

## 数据结构

### MemoResponse

| 字段         | 类型           | 可空 | 说明                                    |
| ------------ | -------------- | ---- | --------------------------------------- |
| `uuid`       | string         | 否   | UUIDv7                                  |
| `content`    | string         | 否   | 已清洗正文                              |
| `row_status` | string         | 否   | `normal` / `archived`                   |
| `tags`       | string[]       | 否   | 独立标签                                |
| `resources`  | ResourceMeta[] | 是   | 资源详情（需 `include_resources=true`） |
| `expires_at` | string         | 是   | ISO8601                                 |
| `headline`   | string         | 是   | 搜索高亮片段                            |
| `created_at` | string         | 否   | ISO8601                                 |
| `updated_at` | string         | 否   | ISO8601                                 |

```json
{
  "uuid": "019e1506-e5ea-7d17-adc8-0785aa7e898f",
  "content": "Meeting notes",
  "row_status": "normal",
  "tags": ["Work", "Ops"],
  "resources": null,
  "expires_at": null,
  "headline": "Meeting <mark>notes</mark>",
  "created_at": "2026-03-01T15:00:00Z",
  "updated_at": "2026-03-01T15:10:00Z"
}
```

### ResourceMeta

| 字段         | 类型   | 说明       |
| ------------ | ------ | ---------- |
| `id`         | string | SHA256     |
| `filename`   | string | 原始文件名 |
| `size`       | int    | 字节数     |
| `mime_type`  | string | MIME       |
| `created_at` | string | ISO8601    |

### HistoryEntry

| 字段           | 类型     | 说明      |
| -------------- | -------- | --------- |
| `id`           | string   | UUIDv7    |
| `memo_uuid`    | string   | 关联 Memo |
| `content`      | string   | 历史正文  |
| `tags`         | string[] | 历史标签  |
| `resource_ids` | string[] | 历史资源  |
| `created_at`   | string   | ISO8601   |
