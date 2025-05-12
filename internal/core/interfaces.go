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

	KV(ctx context.Context, bucket string) (KeyValueClient, error)
}

type BlueskySubscriber interface {
	C() <-chan pips.D[*BlueskyEvent]
}

type Forwarder interface{}

type EventRepository interface {
	Insert(context.Context, ...EventModel) error
}

type AccountRepository interface {
	Insert(context.Context, ...AccountModel) error
	ExistsByDID(context.Context, ...string) ([]string, error)
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
	EstimatedCount(string) (int64, error)

	Migrate() error
}

// KeyValueClient defines the interface for interacting with a JetStream key-value store
type KeyValueClient interface {
	// Get retrieves a value for the given key
	Get(ctx context.Context, key string) ([]byte, error)

	// Put stores a value for the given key
	Put(ctx context.Context, key string, value []byte) error

	// Delete removes a key-value pair
	Delete(ctx context.Context, key string) error

	// Keys returns all keys in the bucket
	Keys(ctx context.Context) ([]string, error)

	// Shutdown gracefully shuts down the client
	Shutdown() error
}
