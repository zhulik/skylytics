package persistence

import "errors"

var (
	ErrNoPostgresDSN = errors.New("no POSTGRESQL_DSN env provided")
)
