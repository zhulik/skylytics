package core

import (
	"time"
)

// PostInteraction stores an interaction with a posts.
type PostInteraction struct {
	ID uint `gorm:"primarykey"`

	CID       string    `gorm:"column:cid"`
	DID       string    `gorm:"column:did"`
	Type      string    `gorm:"column:type"`
	Timestamp time.Time `gorm:"column:timestamp"`
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
