package stormy

import (
	"time"

	"resty.dev/v3"
)

type ClientConfig struct {
	TransportSettings *resty.TransportSettings

	ResponseMiddlewares []resty.ResponseMiddleware
	RequestMiddlewares  []resty.RequestMiddleware
}

var DefaultConfig = &ClientConfig{
	TransportSettings: &resty.TransportSettings{
		DialerTimeout:         1 * time.Second,
		DialerKeepAlive:       1 * time.Second,
		IdleConnTimeout:       1 * time.Second,
		TLSHandshakeTimeout:   1 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 1 * time.Second,
	},
}
