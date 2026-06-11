# 标签集 (Tag Sets)

预定义的筛选条件组合，按分组组织。每个标签集包含：

- `tags_any`: 任一命中（OR）
- `tags_all`: 全部命中（AND）
- `tags_exclude`: 排除命中（NOT）

启动时自动为管理员创建 3 组默认标签集（工作流 / 阅读管理 / 项目追踪）。

---

## 分组管理

### GET /tag-set-groups

列出所有分组。

**Request**

```
GET /api/v1/tag-set-groups
```

**Response 200**

```json
[
  {
    "id": "019e1507-0000-7d17-adc8-0785aa7e898f",
    "name": "工作流",
    "weight": 0,
    "created_at": "2026-03-01T12:00:00Z",
    "updated_at": "2026-03-01T12:00:00Z"
  }
]
```

---

### POST /tag-set-groups

创建分组。

| 参数     | 类型   | 必需 | 默认 | 说明     |
| -------- | ------ | ---- | ---- | -------- |
| `name`   | string | 是   | -    | 分组名称 |
| `weight` | int    | 否   | 0    | 排序权重 |

```json
{ "name": "项目管理", "weight": 0 }
```

**Response 201**: TagSetGroupResponse。

---

### PATCH /tag-set-groups/:id

更新分组。

| 参数     | 类型   | 必需 | 说明   |
| -------- | ------ | ---- | ------ |
| `name`   | string | 否   | 新名称 |
| `weight` | int    | 否   | 新权重 |

```json
{ "name": "新分组名" }
```

**Response 200**: TagSetGroupResponse。

---

### DELETE /tag-set-groups/:id

删除分组。级联删除组内所有标签集。

**Request**

```
DELETE /api/v1/tag-set-groups/:id
```

**Response 204**

---

## 标签集管理

### GET /tag-sets

列出标签集。不传 `group_id` 返回全部。

| 参数       | 类型   | 必需 | 说明     |
| ---------- | ------ | ---- | -------- |
| `group_id` | string | 否   | 按组过滤 |

**Request**

```
GET /api/v1/tag-sets?group_id=group-uuid
```

**Response 200**

```json
[
  {
    "id": "019e1507-1111-7d17-adc8-0785aa7e898f",
    "group_id": "019e1507-0000-7d17-adc8-0785aa7e898f",
    "name": "待办事项",
    "tags_any": ["todo"],
    "tags_all": [],
    "tags_exclude": [],
    "weight": 0,
    "last_used_at": null,
    "created_at": "2026-03-01T12:00:00Z",
    "updated_at": "2026-03-01T12:00:00Z"
  }
]
```

---

### POST /tag-sets

创建标签集。

| 参数           | 类型     | 必需 | 默认 | 说明     |
| -------------- | -------- | ---- | ---- | -------- |
| `name`         | string   | 是   | -    | 名称     |
| `group_id`     | string   | 否   | null | 所属分组 |
| `tags_any`     | string[] | 否   | `[]` | OR 标签  |
| `tags_all`     | string[] | 否   | `[]` | AND 标签 |
| `tags_exclude` | string[] | 否   | `[]` | NOT 标签 |
| `weight`       | int      | 否   | 0    | 排序权重 |

```json
{
  "name": "待办事项",
  "group_id": "019e1507-0000-7d17-adc8-0785aa7e898f",
  "tags_any": ["todo"],
  "tags_all": [],
  "tags_exclude": []
}
```

**Response 201**: TagSetResponse。

---

### GET /tag-sets/:id

获取单个标签集详情。

**Request**

```
GET /api/v1/tag-sets/:id
```

**Response 200**: TagSetResponse。

---

### PATCH /tag-sets/:id

更新标签集。

| 参数           | 类型     | 必需 | 说明          |
| -------------- | -------- | ---- | ------------- |
| `name`         | string   | 否   | 新名称        |
| `group_id`     | string   | 否   | null 移出分组 |
| `tags_any`     | string[] | 否   | 新 OR 标签    |
| `tags_all`     | string[] | 否   | 新 AND 标签   |
| `tags_exclude` | string[] | 否   | 新 NOT 标签   |
| `weight`       | int      | 否   | 新权重        |

**Response 200**: TagSetResponse。

---

### DELETE /tag-sets/:id

删除标签集。

**Request**

```
DELETE /api/v1/tag-sets/:id
```

**Response 204**

---

### POST /tag-sets/:id/touch

更新最后使用时间。

**Request**

```
POST /api/v1/tag-sets/:id/touch
```

**Response 200**: `{}`

---

## 数据结构

### TagSetGroupResponse

| 字段         | 类型   | 说明     |
| ------------ | ------ | -------- |
| `id`         | string | UUIDv7   |
| `name`       | string | 组名     |
| `weight`     | int    | 排序权重 |
| `created_at` | string | ISO8601  |
| `updated_at` | string | ISO8601  |

### TagSetResponse

| 字段           | 类型     | 可空 | 说明     |
| -------------- | -------- | ---- | -------- |
| `id`           | string   | 否   | UUIDv7   |
| `group_id`     | string   | 是   | 所属分组 |
| `name`         | string   | 否   | 名称     |
| `tags_any`     | string[] | 否   | OR       |
| `tags_all`     | string[] | 否   | AND      |
| `tags_exclude` | string[] | 否   | NOT      |
| `weight`       | int      | 否   | 排序     |
| `last_used_at` | string   | 是   | 最后使用 |
| `created_at`   | string   | 否   | ISO8601  |
| `updated_at`   | string   | 否   | ISO8601  |
