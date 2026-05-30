package core

import (
	"context"

	"github.com/stephenafamo/bob"
)

type DB = bob.Transactor[bob.Tx]

type MetricName string

type MetricsCollector interface {
	Increment(ctx context.Context, name MetricName, tags map[string]string)
}
