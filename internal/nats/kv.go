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

func (c *KV) ExistingKeys(ctx context.Context, keys ...string) ([]string, error) {
	l, err := c.kv.ListKeysFiltered(ctx, keys...)
	if err != nil {
		return nil, err
	}
	defer l.Stop() // nolint:errcheck

	var foundKeys []string

	for {
		select {
		case <-ctx.Done():
			return nil, nil
		case key, ok := <-l.Keys():
			if !ok {
				return foundKeys, nil
			}
			foundKeys = append(foundKeys, key)
		}
	}
}
