package core

import (
	"net"
)

type MetricsServer interface {
	Serve(l net.Listener) error
}
