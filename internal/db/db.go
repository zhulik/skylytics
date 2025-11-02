package db

import (
	"context"
	"skylytics/internal/core"

	"github.com/stephenafamo/bob"

	_ "github.com/lib/pq"
)

type DB struct {
	bob.DB
	Config *core.Config
}

func (d *DB) Init(_ context.Context) error {
	db, err := bob.Open("postgres", d.Config.DATABASE_URL)
	d.DB = db

	return err
}
