package core

import (
	"context"

	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/stephenafamo/bob"
)

type BlueskySubscriber interface {
	Consume(ctx context.Context) (chan *models.Event, error)
}

type DB = bob.Transactor[bob.Tx]
