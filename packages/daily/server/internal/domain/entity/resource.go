package entity

import (
	"time"
)

// Resource 是笔记关联的附件
type Resource struct {
	ID           string
	MemoUUID     string
	FileName     string
	Hash         string // 内容 SHA-256
	Size         int64
	MimeType     string
	InternalPath string // 相对存储路径
	CreatedAt    time.Time
}
