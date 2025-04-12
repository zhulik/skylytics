package core

import "skylytics/pkg/async"

type MetricsServer interface{}

type JetstreamSubscriber interface {
	Chan() <-chan async.Result[JetstreamEvent]
}

type Forwarder interface{}
