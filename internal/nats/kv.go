package nats

import (
	"context"
	"fmt"

	"skylytics/internal/core"

	"github.com/nats-io/nats.go/jetstream"
)

// KV implements the core.KeyValueClient interface
type KV struct {
	kv jetstream.KeyValue
}

// NewKV creates a new JetStream key-value client
func NewKV(ctx context.Context, js core.JetstreamClient, bucket string) (core.KeyValueClient, error) {
	kv, err := js.KeyValue(ctx, bucket)
	if err != nil {
		return nil, err
	}
	return &KV{
		kv: kv,
	}, nil
}

// Get retrieves a value for the given key
func (c *KV) Get(ctx context.Context, key string) ([]byte, error) {
	entry, err := c.kv.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	return entry.Value(), nil
}

// Put stores a value for the given key
func (c *KV) Put(ctx context.Context, key string, value []byte) error {
	_, err := c.kv.Put(ctx, key, value)
	if err != nil {
		return fmt.Errorf("failed to store key %s: %w", key, err)
	}
	return nil
}

// Delete removes a key-value pair
func (c *KV) Delete(ctx context.Context, key string) error {
	return c.kv.Delete(ctx, key)
}

// Keys returns all keys in the bucket
func (c *KV) Keys(ctx context.Context) ([]string, error) {
	return c.kv.Keys(ctx)
}

// Shutdown gracefully shuts down the client
func (c *KV) Shutdown() error {
	return nil
}
