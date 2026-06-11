# 运维手册

## 首次初始化

1. 设置 `DAILY_ADMIN_USERNAME` / `DAILY_ADMIN_PASSWORD`
2. 启动服务 → 自动创建管理员账号 + 数据库迁移
3. 可选：`DAILY_BOOTSTRAP_DEMO_MEMOS=true` 导入示例数据
4. 访问 `/api/v1/auth/login` 验证

## 备份与恢复

### 备份

```
# SQLite
cp ./data/daily.db ./backup/daily-$(date +%Y%m%d).db

# PostgreSQL
pg_dump -d daily > ./backup/daily-$(date +%Y%m%d).sql

# 资源文件
tar czf ./backup/storage-$(date +%Y%m%d).tar.gz ./data/storage/
```

### 恢复

- 停止服务 → 替换数据库文件 → 恢复资源目录 → 重启
- 同时备份 `data/daily.db` 和 `data/storage/`（两者一致才完整）

## 迁移发布

- 迁移仅向前（Forward-Only），不支持向下回退
- 新迁移先在小批量数据测试
- 发布前通过 `pre-release checklist` 验证

## 常见故障

| 现象         | 排查                                                 |
| ------------ | ---------------------------------------------------- |
| 登录返回 401 | 检查 Cookie `sid` 是否存在、是否过期                 |
| 导入失败     | 检查 ZIP 结构是否符合导出格式                        |
| 邀请码无效   | 检查 `status` 是否为 `active`、`expires_at` 是否过期 |
| 资源无法访问 | 检查 `owner_user_id` 是否匹配当前用户                |
