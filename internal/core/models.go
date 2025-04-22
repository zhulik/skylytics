package core

import (
	"gorm.io/gorm"
)

// Event represents a raw bluesky jetstream event.
type Event struct {
	gorm.Model

	DID       string       `gorm:"index;type:VARCHAR(32)"`
	Timestamp int64        `gorm:"index:,expression:timestamp DESC"`
	Kind      string       `gorm:"type:VARCHAR(16);index"`
	Event     BlueskyEvent `gorm:"type:jsonb"`
}
