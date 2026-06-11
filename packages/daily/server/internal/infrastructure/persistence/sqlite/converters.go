package sqlite

import (
	"database/sql"
	"time"
)

// toSlTextNull converts a string to sql.NullString
func toSlTextNull(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// fromSlTextNull converts sql.NullString to string
func fromSlTextNull(s sql.NullString) string {
	if !s.Valid {
		return ""
	}
	return s.String
}

// toSlInt8Ptr converts a nullable int64 pointer to sql.NullInt64
func toSlInt8Ptr(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

// fromSlInt8Ptr converts sql.NullInt64 to an int64 pointer
func fromSlInt8Ptr(i sql.NullInt64) *int64 {
	if !i.Valid {
		return nil
	}
	val := i.Int64
	return &val
}

// toSlBool converts a bool to int64 (SQLite doesn't have native bool)
func toSlBool(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// fromSlBool converts int64 to bool
func fromSlBool(i int64) bool {
	return i != 0
}

// toSlTime converts *time.Time to sql.NullTime
func toSlTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

// fromSlTime converts sql.NullTime to *time.Time
func fromSlTime(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	val := t.Time
	return &val
}

