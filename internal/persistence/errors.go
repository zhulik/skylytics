package persistence

import "errors"

var (
	ErrNoPostgresURI = errors.New("no PostgresURI env provided")
)
