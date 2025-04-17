package stormy

import (
	"context"
	"resty.dev/v3"
	"time"
)

const (
	baseURL = "https://public.api.bsky.app"
)

type Client struct {
	client *resty.Client
}

func NewClient() *Client {
	client := resty.NewWithTransportSettings(&resty.TransportSettings{
		DialerTimeout:         1 * time.Second,
		DialerKeepAlive:       1 * time.Second,
		IdleConnTimeout:       1 * time.Second,
		TLSHandshakeTimeout:   1 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 1 * time.Second,
	})

	return &Client{
		client: client,
	}
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) r(ctx context.Context) *resty.Request {
	return c.client.R().WithContext(ctx)
}
