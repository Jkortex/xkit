package tokenizer

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-ego/gse"
)

// GseTokenizer 是基于 gse 实现的分词器
type GseTokenizer struct {
	segmenter gse.Segmenter
	mu        sync.RWMutex
}

func NewGseTokenizer() (*GseTokenizer, error) {
	t := &GseTokenizer{}
	if err := t.Reload(); err != nil {
		return nil, err
	}
	return t, nil
}

func (t *GseTokenizer) Reload() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	s := gse.Segmenter{}

	// 1. 加载默认字典
	dictPath := os.Getenv("GSE_DICT_PATH")
	if dictPath != "" {
		if err := s.LoadDict(dictPath); err != nil {
			return err
		}
	} else {
		if err := s.LoadDict(); err != nil {
			return err
		}
	}

	// 2. 尝试加载用户自定义词典 (data/dict.txt)
	userDict := filepath.Join("data", "dict.txt")
	if _, err := os.Stat(userDict); err == nil {
		_ = s.LoadDict(userDict)
	}

	// 3. 尝试加载停用词 (data/stop_words.txt)
	stopWords := filepath.Join("data", "stop_words.txt")
	if _, err := os.Stat(stopWords); err == nil {
		_ = s.LoadStop(stopWords)
	}

	t.segmenter = s
	return nil
}

func (t *GseTokenizer) Tokenize(text string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// 1. 使用搜索模式切分
	segments := t.segmenter.CutSearch(text, true)

	// 2. 显式过滤停用词
	results := make([]string, 0, len(segments))
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		if t.segmenter.IsStop(seg) {
			continue
		}
		results = append(results, seg)
	}

	return strings.Join(results, " ")
}
