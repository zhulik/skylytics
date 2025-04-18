package stormy

import (
	"context"
	"resty.dev/v3"
)

const (
	baseURL = "https://public.api.bsky.app"
)

type Client struct {
	client *resty.Client
}

func NewClient(config *ClientConfig) *Client {
	if config == nil {
		config = DefaultConfig
	}
	client := resty.NewWithTransportSettings(config.TransportSettings)

	for _, middleware := range config.RequestMiddlewares {
		client.AddRequestMiddleware(middleware)
	}

	for _, middleware := range config.ResponseMiddlewares {
		client.AddResponseMiddleware(middleware)
	}

	return &Client{
		client: client.SetBaseURL(baseURL),
	}
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) r(ctx context.Context) *resty.Request {
	return c.client.R().WithContext(ctx)
}
