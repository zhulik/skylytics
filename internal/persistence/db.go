package persistence

import (
	"context"

	"gorm.io/gorm/logger"

	"skylytics/internal/core"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	Config *core.Config

	*gorm.DB
}

func (db *DB) EstimatedCount(tableName string) (int64, error) {
	var count int64
	return count, db.Raw(
		`SELECT reltuples::bigint AS count 
				FROM pg_class 
				WHERE relname = ?`, tableName,
	).Scan(&count).Error
}

func (db *DB) Init(_ context.Context) error {
	gormDB, err := gorm.Open(postgres.Open(db.Config.PostgresDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}

	db.DB = gormDB

	return nil
}

func (db *DB) Shutdown(_ context.Context) error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return nil
	}
	return sqlDB.Close()
}

func (db *DB) Migrate() error {
	return db.Transaction(func(tx *gorm.DB) error {
		err := tx.AutoMigrate()
		if err != nil {
			return err
		}

		return nil
	})
}
