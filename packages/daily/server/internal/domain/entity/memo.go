package entity

import (
	"time"

	"github.com/google/uuid"
)

type RowStatus string

const (
	RowStatusNormal   RowStatus = "normal"
	RowStatusArchived RowStatus = "archived"
)

const (
	// EphemeralTag 标记为临时笔记的标签
	EphemeralTag = "temp"
	// DefaultEphemeralTTL 临时笔记默认生命周期
	DefaultEphemeralTTL = "3d"
)

// Memo 是系统的核心聚合根
type Memo struct {
	UUID      string    // UUIDv7 唯一标识
	Content   string    // 原始 Markdown 内容
	RowStatus RowStatus // normal, archived
	Tags      []string  // 从内容提取或手动添加的标签
	Resources []*Resource
	ExpiresAt *time.Time // 过期时间，过期后自动归档
	CreatedAt time.Time
	UpdatedAt time.Time
}

// DailyStat 代表某一天的笔记统计数据
type DailyStat struct {
	Date  string
	Count int
}

// TagStat 代表标签及其计数的统计数据
type TagStat struct {
	Name  string
	Count int
}

// NewMemo 创建一个新的笔记实例
func NewMemo(content string) *Memo {
	return &Memo{
		UUID:      uuid.Must(uuid.NewV7()).String(),
		Content:   content,
		RowStatus: RowStatusNormal,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Archive 归档笔记
func (m *Memo) Archive() {
	m.RowStatus = RowStatusArchived
	m.UpdatedAt = time.Now()
}
