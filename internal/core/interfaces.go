package core

import "skylytics/pkg/async"

type MetricsServer interface{}

type BlueskySubscriber interface {
	Chan() <-chan async.Result[BlueskyEvent]
}

type Forwarder interface{}

type CommitAnalyzer interface{}
