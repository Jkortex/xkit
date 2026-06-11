# Daily 后端服务

Daily 是一个专注于个人碎片化知识管理的本地化工具，采用 Clean Architecture 架构，基于 SQLite 提供轻量、零依赖的本地首选（local-first）存储体验。

## 快速入门

- **技术栈**: Go, Gin, SQLite (FTS5), sqlc, gse (分词器)。
- **构建**: `go build -o daily ./cmd/daily/`
- **运行**: `./daily` (自动在本地初始化 `daily.db`)

## 文档指引

所有文档位于 `docs/` 目录：

| 文档                             | 说明                                 |
| -------------------------------- | ------------------------------------ |
| [概览 →](docs/README.md)         | 文档索引                             |
| [架构 →](docs/overview.md)       | Clean Architecture、目录结构、技术栈 |
| [API →](docs/api.md)             | RESTful API v1 参考                  |
| [配置 →](docs/configuration.md)  | 环境变量、本地存储配置               |
| [数据模型 →](docs/data-model.md) | SQL-First、迁移、导入导出            |
| [ADR →](docs/decisions.md)       | 架构决策记录                         |

## 核心特性

- **Zero-Dependency**: 使用 SQLite 作为唯一存储引擎，无需安装/运行外部数据库服务，非常适合本地化及嵌入式部署。
- **Clean Architecture**: 严格的依赖方向，业务逻辑与底层的 SQLite 存储细节解耦。
- **SQL-First**: 使用 `sqlc` 确保 SQLite 查询执行的高效率与类型安全。
- **搜索增强**: SQLite FTS5 配合 gse 分词器，提供极佳的本地搜索体验。
