# 全文检索

## 架构

```
内容 → gse 分词 → tsvector (PG) / FTS5 (SQLite) → GIN 索引 → 检索
```

## 分词

- 使用 [gse](https://github.com/go-ego/gse) 中文分词库
- 索引时对正文、标签、文件名分别分词
- 检索时将用户输入分词后转为 AND 查询

## 存储后端差异

| 能力     | PostgreSQL     | SQLite      |
| -------- | -------------- | ----------- |
| 索引类型 | tsvector + GIN | FTS5 虚拟表 |
| 中文分词 | gse → tsvector | gse → FTS5  |
| 排序     | ts_rank        | bm25        |

## 检索参数（GET /memos）

| 参数               | 说明                                                     |
| ------------------ | -------------------------------------------------------- |
| `search`           | 全文检索词，自动分词 AND 匹配                            |
| `tag`              | 标签精确匹配                                             |
| `tags_any`         | 多标签任一命中（逗号分隔）                               |
| `tags_all`         | 多标签全部命中（逗号分隔）                               |
| `from` / `to`      | 创建时间范围（YYYY-MM-DD）                               |
| `has_resource`     | 是否有附件                                               |
| `sort`             | `created_at_desc` / `created_at_asc` / `updated_at_desc` |
| `limit` / `offset` | 分页                                                     |
