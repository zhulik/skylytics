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
	SaveRaw(ctx context.Context, raw ...[]byte) error
}

type EventsArchiver interface {
	Archive(msg ...jetstream.Msg)
}
