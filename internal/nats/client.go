package nats

import (
	"os"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/do"
)

type Client struct {
	jetstream.JetStream
}

func NewClient(_ *do.Injector) (jetstream.JetStream, error) {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}

	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	c, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	return Client{c}, nil
}

func (c Client) HealthCheck() error {
	return nil
}

func (c Client) Shutdown() error {
	c.Conn().Close()
	return nil
}
