package persistence

import (
	"fmt"
	"os"

	"gorm.io/gorm/logger"

	"skylytics/internal/core"

	"github.com/samber/do"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func (db DB) EstimatedCount(tableName string) (int64, error) {
	var count int64
	return count, db.Raw(
		`SELECT reltuples::bigint AS count 
						FROM pg_class 
						WHERE relname = ?`, tableName,
	).Scan(&count).Error
}

func dsnFromENV() string {
	host := os.Getenv("POSTGRES_HOST")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s", host, user, password, dbname)
}

func NewDB(_ *do.Injector) (core.DB, error) {
	db, err := gorm.Open(postgres.Open(dsnFromENV()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db DB) Migrate() error {
	return db.Transaction(func(tx *gorm.DB) error {
		err := tx.AutoMigrate(&core.EventModel{})
		if err != nil {
			return err
		}

		err = tx.AutoMigrate(&core.AccountModel{})
		if err != nil {
			return err
		}

		return nil
	})
}

func (db DB) LastEventTimestamp() (int64, error) {
	var event core.EventModel
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
