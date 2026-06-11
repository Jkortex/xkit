# 认证与访问控制

## 登录

### POST /auth/login

公开。使用用户名密码登录，验证通过后设置会话 Cookie。

**Request**

```
POST /api/v1/auth/login
Content-Type: application/json
```

| 参数       | 类型   | 必需 | 说明   |
| ---------- | ------ | ---- | ------ |
| `username` | string | 是   | 用户名 |
| `password` | string | 是   | 密码   |

```json
{
  "username": "admin",
  "password": "password123"
}
```

**Response 200**: UserResponse + 设置 `sid` / `remember` Cookie（HttpOnly; Secure; SameSite=Lax）。

**Response 401**

```json
{
  "error": "invalid username or password",
  "code": "UNAUTHORIZED"
}
```

---

## 登出

### POST /auth/logout

清理当前会话 Cookie。

**Request**

```
POST /api/v1/auth/logout
Authorization: Bearer <key> 或 Cookie: sid=<session>
```

**Response 200**: `{}`

---

## 当前用户

### GET /auth/me

返回当前认证用户信息。

**Request**

```
GET /api/v1/auth/me
```

**Response 200**: UserResponse。

**Response 401**

```json
{
  "error": "not authenticated",
  "code": "UNAUTHORIZED"
}
```

---

## 邀请注册

### POST /auth/register-by-invite

公开。使用有效邀请码注册，注册后直接建立登录态。

**Request**

| 参数       | 类型   | 必需 | 说明   |
| ---------- | ------ | ---- | ------ |
| `code`     | string | 是   | 邀请码 |
| `username` | string | 是   | 用户名 |
| `password` | string | 是   | 密码   |

```json
{
  "code": "INVITE-CODE",
  "username": "newuser",
  "password": "password123"
}
```

**Response 200**: UserResponse + Cookie。

**Response 400**

```json
{
  "error": "invite code is invalid or expired",
  "code": "INVALID_INPUT"
}
```

---

## 校验邀请码

### GET /auth/invites/:code/verify

公开。预校验邀请码状态。

**Request**

```
GET /api/v1/auth/invites/:code/verify
```

| 参数   | 位置 | 说明         |
| ------ | ---- | ------------ |
| `code` | path | 邀请码字符串 |

**Response 200**

```json
{
  "valid": true,
  "role": "member",
  "expires_at": "2026-04-01T12:00:00Z"
}
```

---

## 邀请码治理（管理员）

### POST /auth/invites

创建邀请码。响应包含一次性明文 `code`，仅返回一次。

**Request**

| 参数        | 类型   | 必需 | 默认 | 说明               |
| ----------- | ------ | ---- | ---- | ------------------ |
| `role`      | string | 是   | -    | `admin` / `member` |
| `ttl_hours` | int    | 否   | 24   | 有效时长（小时）   |

```json
{ "role": "member", "ttl_hours": 24 }
```

**Response 201**

```json
{
  "id": "019e1506-e5ea-7d17-adc8-0785aa7e898f",
  "code": "INVITE-CODE",
  "role": "member",
  "status": "active",
  "expires_at": "2026-03-02T12:00:00Z",
  "created_at": "2026-03-01T12:00:00Z"
}
```

**Response 403**: 非管理员。

---

### GET /auth/invites

查询邀请码列表。

| 参数     | 类型   | 必需 | 默认 | 说明                                            |
| -------- | ------ | ---- | ---- | ----------------------------------------------- |
| `status` | string | 否   | -    | 过滤：`active` / `used` / `revoked` / `expired` |
| `limit`  | int    | 否   | 100  | 分页大小                                        |
| `offset` | int    | 否   | 0    | 偏移量                                          |

**Response 200**: invite 数组（不含 `code`）。

---

### GET /auth/invites/summary

返回配额摘要。

**Request**

```
GET /api/v1/auth/invites/summary
```

**Response 200**

```json
{
  "active_member_count": 5,
  "active_admin_count": 1,
  "member_limit": 10,
  "admin_limit": 2
}
```

---

### POST /auth/invites/:id/revoke

撤销邀请码。仅 `active` 状态可撤销。

**Request**

```
POST /api/v1/auth/invites/:id/revoke
```

| 参数 | 位置 | 说明        |
| ---- | ---- | ----------- |
| `id` | path | 邀请码 UUID |

**Response 200**: `{}`

**Response 400**

```json
{
  "error": "invite is not active",
  "code": "INVALID_INPUT"
}
```

---

## API Key 管理

### POST /auth/api-keys

为当前用户创建 API Key。明文 key 仅在创建响应返回一次。

**Request**

| 参数        | 类型   | 必需 | 默认 | 说明             |
| ----------- | ------ | ---- | ---- | ---------------- |
| `label`     | string | 是   | -    | 标识             |
| `ttl_hours` | int    | 否   | null | 过期时间（小时） |

```json
{ "label": "CLI Key", "ttl_hours": 720 }
```

**Response 201**

```json
{
  "id": "019e1506-e5ea-7d17-adc8-0785aa7e898f",
  "key": "daily_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "label": "CLI Key",
  "created_at": "2026-03-01T15:00:00Z",
  "expires_at": "2026-03-31T15:00:00Z"
}
```

---

### POST /auth/api-keys/direct

公开。使用用户名密码直接创建 API Key，无需先行登录。

**Request**

| 参数        | 类型   | 必需 | 默认 | 说明             |
| ----------- | ------ | ---- | ---- | ---------------- |
| `username`  | string | 是   | -    | 用户名           |
| `password`  | string | 是   | -    | 密码             |
| `label`     | string | 是   | -    | 标识             |
| `ttl_hours` | int    | 否   | null | 过期时间（小时） |

```json
{
  "username": "admin",
  "password": "password123",
  "label": "CLI Key",
  "ttl_hours": 720
}
```

**Response 201**: 同 `POST /auth/api-keys`。

**Response 401**: 用户名或密码错误。

---

### GET /auth/api-keys

返回当前用户所有 API Key 元数据（不含明文）。

**Request**

```
GET /api/v1/auth/api-keys
```

**Response 200**

```json
[
  {
    "id": "019e1506-e5ea-7d17-adc8-0785aa7e898f",
    "label": "CLI Key",
    "last_used_at": "2026-03-15T10:00:00Z",
    "created_at": "2026-03-01T15:00:00Z",
    "expires_at": "2026-03-31T15:00:00Z"
  }
]
```

---

### DELETE /auth/api-keys/:id

删除指定 API Key。

**Request**

```
DELETE /api/v1/auth/api-keys/:id
```

| 参数 | 位置 | 说明         |
| ---- | ---- | ------------ |
| `id` | path | API Key UUID |

**Response 204**: 无 body。

**Response 404**: Key 不存在。

---

### POST /auth/api-keys/revoke

撤销当前请求所用的 API Key（自注销）。

**Request**

```
POST /api/v1/auth/api-keys/revoke
Authorization: Bearer <key>
```

**Response 200**: `{}`
