package core

import (
	"gorm.io/gorm"
)

// EventModel represents a raw bluesky jetstream event.
type EventModel struct {
	gorm.Model

	Event BlueskyEvent `gorm:"type:jsonb;index:idx_event_did,expression:((event->>'did'));index:idx_event_kind,expression:((event->>'kind'));index:idx_event_time_us,expression:((event->>'time_us'))"`
}

func (EventModel) TableName() string {
	return "events"
}

// AccountModel represents a raw bluesky account event.
type AccountModel struct {
	gorm.Model

	Account []byte `gorm:"type:jsonb;index:idx_did,expression:((account->>'did')),unique"`
}

func (AccountModel) TableName() string {
	return "accounts"
}
