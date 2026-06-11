# Daily Server API v1

Base URL: `/api/v1`
Content-Type: `application/json`（资源上传为 `multipart/form-data`）

---

## 认证

Cookie Session（`sid` Cookie）或 `Authorization: Bearer <key>`。

Bearer 优先级高于 Cookie。若 `Authorization` 存在，直接使用且失败后不回退到 Cookie。

---

## 文档结构

| 文件          | 内容                                                 |
| ------------- | ---------------------------------------------------- |
| `auth.md`     | 登录/登出/当前用户、邀请码注册/治理、API Key 管理    |
| `memo.md`     | Memo CRUD、检索参数、标签提取规则、TTL、历史与回滚   |
| `tag.md`      | 标签治理（统计、重命名、合并、别名、审计）、实现边界 |
| `tag-set.md`  | 标签集分组与标签集管理                               |
| `resource.md` | 资源上传与访问                                       |
| `system.md`   | 统计、数据导出/导入                                  |

---

## 数据结构

### UserResponse

| 字段         | 类型   | 说明               |
| ------------ | ------ | ------------------ |
| `id`         | int    | 用户 ID            |
| `username`   | string | 用户名             |
| `role`       | string | `admin` / `member` |
| `created_at` | string | ISO8601 注册时间   |

```json
{
  "id": 1,
  "username": "admin",
  "role": "admin",
  "created_at": "2026-03-01T12:00:00Z"
}
```

---

## 错误规约

所有错误响应使用统一格式：

```json
{
  "error": "人类可读的错误描述",
  "code": "ERROR_CODE"
}
```

### 错误码

| Code             | HTTP | 说明                               |
| ---------------- | ---- | ---------------------------------- |
| `INVALID_INPUT`  | 400  | 参数非法、校验失败                 |
| `UNAUTHORIZED`   | 401  | 未登录、会话失效、API Key 无效     |
| `FORBIDDEN`      | 403  | 权限不足（如非管理员访问管理接口） |
| `NOT_FOUND`      | 404  | 实体不存在                         |
| `CONFLICT`       | 409  | 业务冲突（如重复创建）             |
| `INTERNAL_ERROR` | 500  | 服务端内部错误                     |

### 追踪

所有响应包含 `X-Request-Id` 头（UUIDv7），跟随请求日志，用于排查。
