package persistence

import (
	"context"
	"database/sql"

	"gorm.io/gorm/logger"

	"skylytics/internal/core"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	db     *gorm.DB
	Config *core.Config
}

func (db *DB) Model(a any) *gorm.DB {
	return db.db.Model(a)
}

func (db *DB) EstimatedCount(tableName string) (int64, error) {
	var count int64
	return count, db.db.Raw(
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

	db.db = gormDB

	return nil
}

func (db *DB) DB() (*sql.DB, error) {
	return db.db.DB()
}

func (db *DB) Shutdown(_ context.Context) error {
	sqlDB, err := db.db.DB()
	if err != nil {
		return nil
	}
	return sqlDB.Close()
}
