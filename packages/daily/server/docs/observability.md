# 可观测性

## 结构化日志

- 使用 Go 1.21+ `log/slog`
- 所有请求日志包含: `request_id`, `user_id`, `duration_ms`, `status`
- 错误日志包含堆栈追踪

## 错误规约

```
400 Bad Request    → INVALID_INPUT    参数非法
401 Unauthorized   → UNAUTHORIZED     未登录/会话失效
403 Forbidden      → FORBIDDEN        权限不足
404 Not Found      → NOT_FOUND        实体不存在
409 Conflict       → CONFLICT         业务冲突
500 Internal       → INTERNAL_ERROR   服务端内部错误
```

响应结构统一:

```json
{ "error": "...", "code": "..." }
```

## 请求追踪

- 每个请求分配 `request_id`（UUIDv7），透传至日志
- `X-Request-Id` 响应头返回

## 指标

- 慢查询日志（>100ms）
- `GET /memos` 查询维度日志（筛选维度、结果量、耗时）
