package db

import (
	"context"
	"skylytics/internal/core"

	"github.com/stephenafamo/bob"

	_ "github.com/jackc/pgx/v5/stdlib" // driver
)

type DB struct {
	bob.DB
	Config *core.Config
}

func (d *DB) Init(_ context.Context) error {
	db, err := bob.Open("pgx", d.Config.DatabaseURL)
	d.DB = db

	return err
}
