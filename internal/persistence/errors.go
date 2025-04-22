package persistence

import "errors"

var (
	PostgresURI = errors.New("no PostgresURI env provided")
)
