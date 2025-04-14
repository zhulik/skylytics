package core

import (
	"context"

	"skylytics/pkg/async"
)

type MetricsServer interface{}

type BlueskySubscriber interface {
	Chan() <-chan async.Result[BlueskyEvent]
}

type Forwarder interface{}

type CommitAnalyzer interface{}

type EventRepository interface {
	SaveRaw(ctx context.Context, raw []byte) error
}
