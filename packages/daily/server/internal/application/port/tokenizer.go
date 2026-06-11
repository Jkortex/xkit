package port

// Tokenizer 定义了分词服务的抽象契约
// 领域层只关心“如何分词”，而不关心是用 gse 还是 jieba 实现
type Tokenizer interface {
	Tokenize(text string) string
	Reload() error // 重新加载词典与停用词
}
