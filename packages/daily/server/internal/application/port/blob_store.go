package port

import (
	"context"
	"io"
)

// BlobStore 定义了二进制大对象的存储契约
type BlobStore interface {
	// Put 保存文件，返回相对路径和可能的错误
	Put(ctx context.Context, relPath string, reader io.Reader) error
	// Get 读取文件
	Get(ctx context.Context, relPath string) (io.ReadCloser, error)
	// Delete 删除文件
	Delete(ctx context.Context, relPath string) error
	// ListAll 遍历并返回所有文件的相对路径
	ListAll(ctx context.Context) ([]string, error)
}
