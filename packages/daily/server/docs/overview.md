# 架构总览

## 核心设计哲学

- **Clean Architecture**: 业务逻辑与外部技术栈彻底解耦。依赖方向严格单向。
- **Interface Adapter Pattern**: Handler → Controller → Presenter 三段式适配。
- **SQL-First**: 通过 `sqlc` 保持 100% SQL 控制权，无 ORM。

## 目录结构

```
internal/
├── domain/              # [核心层] 实体、领域服务（零外部依赖）
│   ├── entity/           Memo, Tag, Resource, User 等
│   └── service/          TagExtractor 等纯逻辑服务
├── application/         # [应用层] 用例编排、端口定义
│   ├── port/             Repository 接口、DTO
│   ├── dto/              请求/响应 DTO
│   └── usecase/          用例实现（MemoService, TagService 等）
├── interfaces/          # [接口适配层] 框架无关契约
│   ├── controller/       参数组装、用例触发
│   └── presenter/        响应呈现接口
└── infrastructure/      # [基础设施层] 具体实现
    ├── api/               Gin Handler、JSON Presenter
    ├── persistence/       sqlc 生成的 SQL 实现（仅 SQLite）
    └── storage/           本地文件系统
```

## 依赖方向

```
Infrastructure → Interfaces → Application → Domain
```

内层不依赖外层。Domain 无任何外部导入。

## 技术栈

| 组件      | 选型                  |
| --------- | --------------------- |
| 语言      | Go 1.25.0             |
| HTTP 框架 | Gin                   |
| 持久化    | SQLite 3 (FTS5)       |
| SQL 生成  | sqlc                  |
| 中文分词  | gse                   |
| 迁移      | 内嵌 SQL 脚本自动迁移 |
| 日志      | slog                  |
