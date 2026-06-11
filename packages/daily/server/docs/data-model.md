# 数据模型

## SQL-First 工作流

```
DDL 迁移 → sqlc 生成 → Go 代码使用
```

1. 在 `cmd/daily/cmd/migrations/` 编写 SQL 迁移
2. 运行 `sqlc generate` 从 SQL 生成类型安全 Go 代码
3. 应用层通过 Repository 接口操作，不直接调用 sqlc

## 核心实体

### Memo（笔记）

| 字段            | 类型     | 说明                             |
| --------------- | -------- | -------------------------------- |
| `memo_uuid`     | TEXT     | 业务主键 (UUIDv7)                |
| `content`       | TEXT     | 清洗后正文（不含尾部标签块）     |
| `row_status`    | TEXT     | `normal` / `archived`            |
| `owner_user_id` | INTEGER  | 所有者用户 ID (默认为本地管理员) |
| `search_text`   | TEXT     | FTS 索引内容                     |
| `expires_at`    | DATETIME | 临时笔记过期时间                 |

### Tag（标签）

- 标签通过 `memo_tag` 关联表与 Memo 多对多关联
- 标签治理操作记录在 `tag_governance_audit` 表

### Resource（资源）

| 字段            | 类型    | 说明            |
| --------------- | ------- | --------------- |
| `id`            | TEXT    | SHA256 内容寻址 |
| `filename`      | TEXT    | 原始文件名      |
| `owner_user_id` | INTEGER | 所有者用户 ID   |
| `mime_type`     | TEXT    | 自动检测        |

## 数据隔离

虽然数据库表结构中依然保留了 `owner_user_id` 列，但在本地化工具使用场景中，所有的操作与 API 调用都将统一在本地启动时自动拉起并配置好的默认管理员账号下安全执行。

## 导入导出

- **导出**: `GET /system/export` → 返回完整 ZIP。先写临时文件，成功后再回传。
- **导入**: `POST /system/import` → 解压 ZIP，按 `memo_uuid` + `content+created_at` 指纹幂等多键去重。
- **幂等**: 重复导入同一归档，`imported` 计数收敛为 0。
