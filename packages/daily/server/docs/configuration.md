# 配置体系

## 优先级（高→低）

1. CLI 参数 `-c config.json`
2. 环境变量 `DAILY_*`
3. 配置文件（默认 `~/.daily/config.json`）
4. 硬编码默认值

## 关键配置项与环境变量

| 配置项 (JSON Key)           | 环境变量                          | 默认值                                 | 说明                           |
| --------------------------- | --------------------------------- | -------------------------------------- | ------------------------------ |
| `sqlite_dsn`                | `DAILY_SQLITE_DSN`                | `./data/daily.db`                      | SQLite 本地数据库文件路径      |
| `storage_dir`               | `DAILY_STORAGE_DIR`               | `./data/storage`                       | 资源/附件文件本地存储目录      |
| `port`                      | `DAILY_PORT`                      | `8080`                                 | HTTP 本地服务监听端口          |
| `bootstrap_admin_username`  | `DAILY_BOOTSTRAP_ADMIN_USERNAME`  | `admin`                                | 首次启动自动创建的管理员用户名 |
| `bootstrap_admin_password`  | `DAILY_BOOTSTRAP_ADMIN_PASSWORD`  | `admin`                                | 管理员账号的初始密码           |
| `bootstrap_demo_memos`      | `DAILY_BOOTSTRAP_DEMO_MEMOS`      | `false`                                | 是否自动导入种子示例数据       |
| `bootstrap_demo_memos_path` | `DAILY_BOOTSTRAP_DEMO_MEMOS_PATH` | `./docs/seeds/architecture_memos.json` | 种子数据路径                   |
| `theme`                     | `TUI_THEME`                       | `ocean`                                | TUI 的配色主题                 |

## 启动验证

启动时自动检查：

- SQLite 数据库可达与可写
- 本地存储目录可写
- 管理员账号存在（不存在则自动创建，后续本地客户端请求直接默认以此账号身份安全运行，无需反复手动登录）
