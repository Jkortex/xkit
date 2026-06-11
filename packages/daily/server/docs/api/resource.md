# 资源管理

## 上传

### POST /resources

上传资源文件。自动检测 MIME 类型。

**Request**

```
POST /api/v1/resources
Content-Type: multipart/form-data
```

| 参数   | 类型 | 必需 | 说明                        |
| ------ | ---- | ---- | --------------------------- |
| `file` | file | 是   | 上传文件（multipart field） |

**Response 201**

| 字段         | 类型   | 说明            |
| ------------ | ------ | --------------- |
| `id`         | string | SHA256 内容寻址 |
| `filename`   | string | 原始文件名      |
| `size`       | int    | 字节数          |
| `mime_type`  | string | 自动检测 MIME   |
| `created_at` | string | ISO8601         |

```json
{
  "id": "abcdef1234567890abcdef1234567890abcdef12",
  "filename": "notes.pdf",
  "size": 102400,
  "mime_type": "application/pdf",
  "created_at": "2026-03-01T12:00:00Z"
}
```

---

## 访问

### GET /resources/:id

返回原始文件流。仅允许访问当前用户资源。

**Request**

```
GET /api/v1/resources/:id
```

| 参数 | 位置 | 说明           |
| ---- | ---- | -------------- |
| `id` | path | 资源 SHA256 ID |

**Response 200**: `Content-Type: application/octet-stream`（或实际 MIME）。

**Response 404**: 不存在或不属于当前用户。
