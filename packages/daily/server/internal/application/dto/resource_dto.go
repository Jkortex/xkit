package dto

import (
	"io"
)

// UploadResourceInput 资源上传的输入
type UploadResourceInput struct {
	FileName string
	MimeType string
	Size     int64
	Content  io.Reader
}

type ImportSkipDetail struct {
	Entity string `json:"entity"`
	Key    string `json:"key"`
	Reason string `json:"reason"`
}

type ImportSectionReport struct {
	Imported int                `json:"imported"`
	Skipped  int                `json:"skipped"`
	Details  []ImportSkipDetail `json:"details"`
}

type ImportReport struct {
	Memos     ImportSectionReport `json:"memos"`
	Resources ImportSectionReport `json:"resources"`
}
