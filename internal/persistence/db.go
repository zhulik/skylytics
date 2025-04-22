package persistence

import (
	"os"

	"skylytics/internal/core"

	"github.com/samber/do"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
// defaultDBDSN = "host=localhost user=bsky password=bsky dbname=bsky sslmode=disable"
)

type DB struct {
	*gorm.DB
}

func NewDB(_ *do.Injector) (core.DB, error) {
	dsn := os.Getenv("POSTGRESQL_DSN")
	if dsn == "" {
		return nil, ErrNoPostgresDSN
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db DB) Migrate() error {
	return db.Transaction(func(tx *gorm.DB) error {
		err := tx.AutoMigrate(&core.Event{})
		if err != nil {
			return err
		}

		return nil
	})
}

func (db DB) LastEventTimestamp() (int64, error) {
	var event core.Event
	err := db.Order("timestamp DESC").First(&event).Error
	if err != nil {
		return 0, err
	}
	return event.Event.TimeUS, nil
}

func ProcessedIs(processed bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("processed = ?", processed)
	}
}

func Processed(db *gorm.DB) *gorm.DB {
	return ProcessedIs(true)(db)
}

func NotProcessed(db *gorm.DB) *gorm.DB {
	return ProcessedIs(false)(db)
}
