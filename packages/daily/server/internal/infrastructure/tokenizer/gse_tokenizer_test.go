package tokenizer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGseTokenizer_Tokenize(t *testing.T) {
	// 1. 准备测试数据目录
	testDataDir := "testdata"
	err := os.MkdirAll(testDataDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDataDir)

	// 2. 创建测试停用词文件
	stopWordsPath := filepath.Join(testDataDir, "stop_words.txt")
	stopWordsContent := "的\n了\n和\n是\n很"
	err = os.WriteFile(stopWordsPath, []byte(stopWordsContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// 3. 创建测试自定义词典
	dictPath := filepath.Join(testDataDir, "dict.txt")
	dictContent := "项目管理 500 n\n知识库 500 n"
	err = os.WriteFile(dictPath, []byte(dictContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// 4. 为了测试，我们需要临时更改工作目录或让 NewGseTokenizer 接受路径
	// 简单起见，我们手动创建一个符合结构的 Tokenizer 用于测试
	// 或者直接在当前目录创建 data 文件夹（测试完删除）
	err = os.MkdirAll("data", 0755)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("data")

	_ = os.WriteFile(filepath.Join("data", "stop_words.txt"), []byte(stopWordsContent), 0644)
	_ = os.WriteFile(filepath.Join("data", "dict.txt"), []byte(dictContent), 0644)

	tokenizer, err := NewGseTokenizer()
	if err != nil {
		t.Fatalf("Failed to create tokenizer: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		contains []string
		excludes []string
	}{
		{
			name:     "Filter stop words",
			input:    "我的会议和项目",
			contains: []string{"我", "会议", "项目"},
			excludes: []string{"的", "和"},
		},
		{
			name:     "User dictionary",
			input:    "我们要做好项目管理",
			contains: []string{"项目管理"},
		},
		{
			name:     "Combined",
			input:    "知识库是很有用的",
			contains: []string{"知识库", "有用"},
			excludes: []string{"是", "很", "的"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tokenizer.Tokenize(tt.input)
			for _, c := range tt.contains {
				if !strings.Contains(got, c) {
					t.Errorf("Tokenize(%q) result %q should contain %q", tt.input, got, c)
				}
			}
			for _, e := range tt.excludes {
				if strings.Contains(got, " "+e+" ") || strings.HasPrefix(got, e+" ") || strings.HasSuffix(got, " "+e) {
					// 精确匹配单词，避免子串误判（如“的”在“项目”中不应被排除，但此处 gse 返回的是空格分隔的词）
					found := false
					words := strings.Fields(got)
					for _, w := range words {
						if w == e {
							found = true
							break
						}
					}
					if found {
						t.Errorf("Tokenize(%q) result %q should NOT contain stop word %q", tt.input, got, e)
					}
				}
			}
		})
	}
}
