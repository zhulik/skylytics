package core

import (
	"net"
)

type MetricsServer interface {
	Serve(l net.Listener) error
}

type JetstreamSubscriber interface {
	Chan() <-chan string
}

type Forwarder interface{}
