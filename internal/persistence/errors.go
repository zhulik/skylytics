package persistence

import "errors"

var (
	ErrNoMongodbURI = errors.New("no MONGODB_URI env provided")
)
