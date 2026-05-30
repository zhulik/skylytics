package core

import (
	"context"
	"time"

	libredis "github.com/redis/go-redis/v9"
	"github.com/stephenafamo/bob"
)

type DB = bob.Transactor[bob.Tx]

type MetricsCollector interface {
	IncJetstreamProcessedEventsTotal(ctx context.Context, kind, operation, collection string)
}

type Redis interface {
	Get(ctx context.Context, key string) *libredis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *libredis.StatusCmd
}
