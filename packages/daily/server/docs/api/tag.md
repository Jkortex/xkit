# 标签治理

## 标签统计

### GET /tags

返回当前用户所有标签及使用计数。

**Request**

```
GET /api/v1/tags
```

**Response 200**

```json
[
  { "name": "Work", "count": 10 },
  { "name": "Ops", "count": 5 }
]
```

---

## 重命名（合并）

### POST /tags/rename

单标签重命名。若 `to` 已存在则执行合并（旧标签关联 Memo 转关联到目标标签）。

**Request**

| 参数   | 类型   | 必需 | 说明     |
| ------ | ------ | ---- | -------- |
| `from` | string | 是   | 源标签   |
| `to`   | string | 是   | 目标标签 |

```json
{ "from": "Legacy", "to": "Ops" }
```

**Response 200**

| 字段             | 类型   | 说明                                |
| ---------------- | ------ | ----------------------------------- |
| `from`           | string | 源标签                              |
| `to`             | string | 目标标签                            |
| `affected_memos` | int    | 影响 Memo 数                        |
| `merged`         | bool   | 是否合并（true）或纯重命名（false） |

```json
{
  "from": "Legacy",
  "to": "Ops",
  "affected_memos": 3,
  "merged": true
}
```

**Response 404**: `from` 不存在。

---

## 批量合并

### POST /tags/merge

多源标签并入一个目标。目标不存在时自动创建。

**Request**

| 参数      | 类型     | 必需 | 说明       |
| --------- | -------- | ---- | ---------- |
| `sources` | string[] | 是   | 源标签列表 |
| `target`  | string   | 是   | 目标标签   |

```json
{ "sources": ["Infra", "Platform"], "target": "Ops" }
```

**Response 200**

| 字段              | 类型     | 说明               |
| ----------------- | -------- | ------------------ |
| `sources`         | string[] | 请求的源标签       |
| `target`          | string   | 目标标签           |
| `affected_memos`  | int      | 影响 Memo 总数     |
| `merged_sources`  | string[] | 实际合并的源       |
| `skipped_sources` | string[] | 跳过（不存在）的源 |

```json
{
  "sources": ["Infra", "Platform"],
  "target": "Ops",
  "affected_memos": 10,
  "merged_sources": ["Infra", "Platform"],
  "skipped_sources": []
}
```

---

## 别名管理

### GET /tags/aliases

返回别名注册表（系统级）。

**Request**

```
GET /api/v1/tags/aliases
```

**Response 200**

```json
[
  { "alias": "SRE", "canonical": "Ops" },
  { "alias": "TODO", "canonical": "Task" }
]
```

---

### POST /tags/aliases

创建/更新别名。新建/更新 Memo 时输入标签自动解析为 canonical。检索 `?tag=<alias>` 自动展开。

**Request**

| 参数        | 类型   | 必需 | 说明   |
| ----------- | ------ | ---- | ------ |
| `alias`     | string | 是   | 别名   |
| `canonical` | string | 是   | 规范名 |

```json
{ "alias": "SRE", "canonical": "Ops" }
```

**Response 200**

```json
{ "alias": "SRE", "canonical": "Ops" }
```

> 若历史上已存在 `alias` 标签数据，创建别名时会尝试将旧标签数据并入 canonical。

---

### DELETE /tags/aliases/:alias

删除别名。

**Request**

```
DELETE /api/v1/tags/aliases/:alias
```

**Response 204**

---

## 审计

### GET /tags/audits

查看治理操作流水。

| 参数     | 类型   | 必需 | 默认 | 说明                                                       |
| -------- | ------ | ---- | ---- | ---------------------------------------------------------- |
| `limit`  | int    | 否   | 20   | 返回条数                                                   |
| `action` | string | 否   | -    | 过滤：`rename` / `merge` / `alias_upsert` / `alias_delete` |

**Request**

```
GET /api/v1/tags/audits?limit=20&action=merge
```

**Response 200**

```json
[
  {
    "action": "merge",
    "summary": "Infra,Platform -> Ops",
    "affected_memos": 10,
    "created_at": "2026-03-01T15:00:00Z"
  }
]
```

---

## 实现边界

- 标签统计与 Memo 数据按当前用户作用域返回
- **`tag_alias` 与 `tag_governance_audit` 为系统级共享表**（无 `user_id`），非严格用户私有配置
