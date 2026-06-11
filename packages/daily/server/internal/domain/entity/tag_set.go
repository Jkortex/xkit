package entity

import "time"

type TagSetGroup struct {
	ID        string
	UserID    int64
	Name      string
	Weight    int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TagSet struct {
	ID          string
	UserID      int64
	GroupID     *string
	Name        string
	TagsAny     string // JSON array
	TagsAll     string // JSON array
	TagsExclude string // JSON array
	Weight      int
	LastUsedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
