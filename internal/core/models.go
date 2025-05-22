package core

import (
	"time"
)

// PostStats stores commit statistics
type PostStats struct {
	CID string `gorm:"primaryKey"`

	Likes   int64
	Reposts int64
	Replies int64

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Post struct {
	Text  *string
	Langs []string
	Reply Reply
}

type Reply struct {
	Parent string
	Root   string
}
