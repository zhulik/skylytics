package core

import (
	"time"

	"gorm.io/gorm"
)

// EventModel represents a raw bluesky jetstream event.
type EventModel struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt

	Event BlueskyEvent
}

func (EventModel) TableName() string {
	return "events"
}

// AccountModel represents a raw bluesky account event.
type AccountModel struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Account []byte `gorm:"type:jsonb;index:idx_did,expression:((account->>'did')),unique"`
}

func (AccountModel) TableName() string {
	return "accounts"
}
