package core

import (
	"context"
	"github.com/nats-io/nats.go/jetstream"

	"skylytics/pkg/async"
)

type MetricsServer interface{}

type BlueskySubscriber interface {
	Chan() <-chan async.Result[BlueskyEvent]
}

type Forwarder interface{}

type CommitAnalyzer interface{}

type EventRepository interface {
	InsertRaw(ctx context.Context, raw ...[]byte) ([]any, error)
}

type AccountRepository interface {
	InsertRaw(ctx context.Context, raw ...[]byte) ([]any, error)
}

type AccountUpdater interface {
	Update(msg ...jetstream.Msg) error
}

type EventsArchiver interface {
	Archive(msg ...jetstream.Msg) error
}
