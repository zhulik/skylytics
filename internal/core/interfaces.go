package core

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"
	"gorm.io/gorm"

	"github.com/zhulik/pips"
)

type MetricsServer interface{}

type JetstreamClient interface {
	jetstream.JetStream

	Consume(ctx context.Context, stream, name string) (<-chan pips.D[jetstream.Msg], error)
	ConsumeToPipeline(ctx context.Context, stream, name string, pipeline *pips.Pipeline[jetstream.Msg, any]) error
}

type BlueskySubscriber interface {
	Subscribe() <-chan pips.D[BlueskyEvent]
}

type Forwarder interface{}

type CommitAnalyzer interface{}

type EventRepository interface {
	Insert(ctx context.Context, raw ...EventModel) error
}

type AccountRepository interface {
	Insert(ctx context.Context, raw ...[]byte) error
	ExistsByDID(ctx context.Context, dids ...string) ([]string, error)
}

type AccountUpdater interface {
	Update(ctx context.Context, msg ...jetstream.Msg) error
}

type EventsArchiver interface {
	Archive(ctx context.Context, msg ...jetstream.Msg) error
}

type MetricsCollector interface{}

type DB interface {
	Model(any) *gorm.DB

	LastEventTimestamp() (int64, error)
	Migrate() error
}
