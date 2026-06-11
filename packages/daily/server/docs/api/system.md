# 系统

## 统计

### GET /stats

返回当前用户的资产总量与年度热力图。

**Request**

```
GET /api/v1/stats
```

**Response 200**

| 字段              | 类型  | 说明                                  |
| ----------------- | ----- | ------------------------------------- |
| `memos_total`     | int   | Memo 总量                             |
| `tags_total`      | int   | 标签去重计数                          |
| `resources_total` | int   | 资源总量                              |
| `heatmap`         | array | `[{date, count}]` 最近 365 天按日聚合 |

```json
{
  "memos_total": 100,
  "tags_total": 20,
  "resources_total": 50,
  "heatmap": [
    { "date": "2026-03-01", "count": 5 },
    { "date": "2026-03-02", "count": 3 }
  ]
}
```

---

## 导出

### GET /system/export

导出当前用户完整数据 ZIP。

**Request**

```
GET /api/v1/system/export
```

**Response 200**: `Content-Type: application/zip`

ZIP 内部结构：

```
memos.json
resources.json
assets/<internal_path>
```

> 先完整生成 ZIP 到临时文件，成功后再整体回传。避免流式写入中断导致损坏包。

**Response 500**: 打包失败。

---

## 导入

### POST /system/import

上传 ZIP 恢复数据。幂等多键去重，可重试、部分成功。

**Request**

```
POST /api/v1/system/import
Content-Type: multipart/form-data
```

| 参数   | 类型 | 必需 | 说明       |
| ------ | ---- | ---- | ---------- |
| `file` | file | 是   | 导出的 ZIP |

**Response 200**

| 字段                 | 类型   | 说明             |
| -------------------- | ------ | ---------------- |
| `memos_imported`     | int    | 成功导入 Memo 数 |
| `resources_imported` | int    | 成功导入资源数   |
| `memos_skipped`      | int    | 跳过 Memo 数     |
| `resources_skipped`  | int    | 跳过资源数       |
| `report`             | object | 结构化明细       |

```json
{
  "memos_imported": 10,
  "resources_imported": 5,
  "memos_skipped": 2,
  "resources_skipped": 0,
  "report": {
    "memos": {
      "imported": 10,
      "skipped": 2,
      "details": [{ "reason": "duplicate_by_memo_uuid", "count": 2 }]
    },
    "resources": {
      "imported": 5,
      "skipped": 0,
      "details": []
    }
  }
}
```

### 去重逻辑（幂等）

- 资源：按 `id` / hash / `internal_path` 多键去重
- Memo：按 `memo_uuid` 去重
- 同一 ZIP 重复导入，imported 计数收敛为 0

### 跳过原因

| reason                   | 说明             |
| ------------------------ | ---------------- |
| `duplicate_by_id`        | 资源 ID 已存在   |
| `duplicate_by_hash`      | 资源哈希已存在   |
| `duplicate_by_path`      | 资源路径已存在   |
| `duplicate_by_memo_uuid` | Memo UUID 已存在 |
| `invalid_metadata`       | 元数据格式非法   |
| `invalid_memo_uuid`      | Memo UUID 非法   |
| `invalid_memo_reference` | 引用资源不存在   |
| `path_traversal_attempt` | 路径穿越攻击     |
