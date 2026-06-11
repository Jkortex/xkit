package dto

import "time"

// --- TagSetGroup ---

type CreateTagSetGroupRequest struct {
	Name   string `json:"name" binding:"required"`
	Weight int    `json:"weight"`
}

type UpdateTagSetGroupRequest struct {
	Name   *string `json:"name"`
	Weight *int    `json:"weight"`
}

type TagSetGroupResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Weight    int       `json:"weight"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// --- TagSet ---

type CreateTagSetRequest struct {
	Name        string   `json:"name" binding:"required"`
	GroupID     *string  `json:"group_id"`
	TagsAny     []string `json:"tags_any"`
	TagsAll     []string `json:"tags_all"`
	TagsExclude []string `json:"tags_exclude"`
	Weight      int      `json:"weight"`
}

type UpdateTagSetRequest struct {
	Name        *string  `json:"name"`
	GroupID     **string `json:"group_id"`
	TagsAny     []string `json:"tags_any"`
	TagsAll     []string `json:"tags_all"`
	TagsExclude []string `json:"tags_exclude"`
	Weight      *int     `json:"weight"`
}

type TagSetResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	GroupID     *string    `json:"group_id"`
	TagsAny     []string   `json:"tags_any"`
	TagsAll     []string   `json:"tags_all"`
	TagsExclude []string   `json:"tags_exclude"`
	Weight      int        `json:"weight"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
