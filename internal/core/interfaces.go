package core

type MetricsServer interface{}

type JetstreamSubscriber interface {
	Chan() <-chan JetstreamEvent
}

type Forwarder interface{}
