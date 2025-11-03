package db

import (
	"context"
	"skylytics/internal/core"

	"github.com/stephenafamo/bob"

	_ "github.com/lib/pq" // driver
)

type DB struct {
	bob.DB
	Config *core.Config
}

func (d *DB) Init(_ context.Context) error {
	db, err := bob.Open("postgres", d.Config.DatabaseURL)
	d.DB = db

	return err
}
