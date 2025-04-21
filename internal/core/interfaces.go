package core

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/zhulik/pips"
)

type MetricsServer interface{}

type JetstreamClient interface {
	jetstream.JetStream

	Consume(ctx context.Context, stream, name string) (<-chan pips.D[jetstream.Msg], error)
	ConsumeToPipeline(ctx context.Context, stream, name string, pipeline *pips.Pipeline[jetstream.Msg, any]) error
}

type BlueskySubscriber interface {
	Chan(ctx context.Context) <-chan pips.D[BlueskyEvent]
}

type Forwarder interface{}

type CommitAnalyzer interface{}

type EventRepository interface {
	InsertRaw(ctx context.Context, raw ...[]byte) ([]any, error)
}

type AccountRepository interface {
	InsertRaw(ctx context.Context, raw ...[]byte) ([]any, error)
	ExistsByDID(ctx context.Context, dids ...string) ([]string, error)
}

type AccountUpdater interface {
	Update(ctx context.Context, msg ...jetstream.Msg) error
}

type EventsArchiver interface {
	Archive(ctx context.Context, msg ...jetstream.Msg) error
}

type MetricsCollector interface{}
