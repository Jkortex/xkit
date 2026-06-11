package entity

import "time"

type UserRole string

const (
	UserRoleAdmin  UserRole = "admin"
	UserRoleMember UserRole = "member"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         UserRole
	Status       UserStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Session struct {
	ID                string
	UserID            int64
	SessionTokenHash  string
	RememberTokenHash string
	ExpiresAt         time.Time
	RememberExpiresAt time.Time
	RevokedAt         *time.Time
	UserAgent         string
	ClientIP          string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Invite struct {
	ID        string
	CodeHash  string
	Role      UserRole
	ExpiresAt time.Time
	UsedAt    *time.Time
	UsedBy    *int64
	CreatedBy *int64
	RevokedAt *time.Time
	CreatedAt time.Time
}

type ApiKey struct {
	ID         string
	UserID     int64
	KeyHash    string
	Label      string
	CreatedAt  time.Time
	ExpiresAt  *time.Time
	LastUsedAt *time.Time
}
