# Daily Server 文档

```
docs/
├── README.md           ← 本文档
├── overview.md         # 架构总览：Clean Architecture、目录结构、技术栈
├── api.md              # REST API v1 参考
├── configuration.md    # 配置体系：环境变量、优先级
├── data-model.md       # 数据模型：SQL-First、sqlc、迁移、导入导出
├── search.md           # 全文检索：GSE 分词、tsvector/GIN、检索 DSL
├── tags.md             # 标签体系：提取算法、标签集、治理（重命名/合并/别名/审计）
├── resources.md        # 资源管理：内容寻址存储、访问控制、GC
├── observability.md    # 可观测性：结构化日志、错误规约、追踪
├── testing.md          # 测试指南：金字塔结构、运行方式
├── ops.md              # 运维手册：部署、备份、故障排查
├── release.md          # 发布检查单
├── capabilities.md     # 业务能力地图
├── decisions.md        # 架构决策记录（ADR）
└── seeds/
    └── architecture_memos.json
```

## 快速链接

| 用途     | 文档                                                                                     |
| -------- | ---------------------------------------------------------------------------------------- |
| 启动服务 | [配置 →](configuration.md) [运维 →](ops.md)                                              |
| 理解架构 | [概览 →](overview.md)                                                                    |
| 开发接口 | [API →](api.md)                                                                          |
| 理解数据 | [数据模型 →](data-model.md) [搜索 →](search.md) [标签 →](tags.md) [资源 →](resources.md) |
| 定位问题 | [可观测性 →](observability.md) [运维 →](ops.md)                                          |
| 发布上线 | [发布检查 →](release.md)                                                                 |
| 背景决策 | [ADR →](decisions.md)                                                                    |
