package core

import (
	"net"
)

type MetricsServer interface {
	Serve(l net.Listener) error
}

type JetstreamSubscriber interface {
	Chan() <-chan JetstreamEvent
}

type Forwarder interface{}
