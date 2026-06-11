# 测试指南

## 测试金字塔

```
┌──────────┐
│  E2E     │  API 集成测试（PostgreSQL）
├──────────┤
│  集成    │  Repository 测试（SQLite 内存模式）
├──────────┤
│  单元    │  Domain Service / UseCase 测试
└──────────┘
```

## 运行方式

| 命令                                 | 范围                           |
| ------------------------------------ | ------------------------------ |
| `make test-sl`                       | 全部测试（SQLite）             |
| `make test-pg`                       | 全部测试（PostgreSQL，需容器） |
| `go test ./internal/domain/...`      | 仅领域层单元测试               |
| `go test ./internal/application/...` | 仅应用层单元测试               |

## 编写规范

- 单元测试不依赖外部资源（DB/网络/文件系统）
- Repository 测试使用 SQLite 内存模式（`:memory:`）
- 集成测试可用 `testcontainers-go` 启动 PostgreSQL
- 基准测试命名 `BenchmarkXxx`，用于监控性能回归
