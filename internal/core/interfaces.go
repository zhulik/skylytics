package core

import (
	"context"
	"database/sql"

	"github.com/nats-io/nats.go/jetstream"
	"gorm.io/gorm"

	"github.com/zhulik/pips"
)

type MetricsServer interface{}

type JetstreamClient interface {
	jetstream.JetStream

	ConsumeToPipeline(ctx context.Context, stream, name string, pipeline *pips.Pipeline[jetstream.Msg, any]) error

	KV(ctx context.Context, bucket string) (KeyValueClient, error)
}

type BlueskySubscriber interface {
	ConsumeToPipeline(ctx context.Context, pipeline *pips.Pipeline[*BlueskyEvent, *BlueskyEvent]) error
}

type Forwarder interface{}

type AccountUpdater interface {
}

type EventsArchiver interface {
	Archive(ctx context.Context, msg ...jetstream.Msg) error
}

type MetricsCollector interface{}
type Migrator interface {
	Up(ctx context.Context) error
	Down(ctx context.Context) error
	Migrate(ctx context.Context, version uint) error
}

type DB interface {
	Model(any) *gorm.DB
	DB() (*sql.DB, error)
	EstimatedCount(string) (int64, error)
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

	// ExistingKeys returns a list of keys that already exist in the bucket.
	ExistingKeys(context.Context, ...string) ([]string, error)
}

type PostRepository interface {
	Get(ctx context.Context, cid string) (Post, error)
	AddInteraction(ctx context.Context, interaction PostInteraction) error
	TopN(ctx context.Context, n int) ([]Post, error)
}
