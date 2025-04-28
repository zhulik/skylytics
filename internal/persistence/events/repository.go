package events

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"skylytics/internal/core"

	"github.com/samber/do"
)

const (
	// KVBucketName is the name of the KV bucket for events
	KVBucketName = "skylytics-events"
)

type Repository struct {
	kv core.KeyValueClient
}

func NewRepository(i *do.Injector) (core.EventRepository, error) {
	js := do.MustInvoke[core.JetstreamClient](i)

	kv, err := js.KV(context.Background(), KVBucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to create KV client: %w", err)
	}

	return Repository{
		kv: kv,
	}, nil
}

func (r Repository) Insert(ctx context.Context, events ...core.EventModel) error {
	for i, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event %d: %w", i, err)
		}

		if err := r.kv.Put(ctx, key(event), data); err != nil {
			return fmt.Errorf("failed to store event %d, %w", i, err)
		}
	}

	return nil
}

func key(event core.EventModel) string {
	hashedDID := sha256.Sum256([]byte(event.Event.Did))
	return fmt.Sprintf("event.%x.%d", hashedDID, event.Event.TimeUS)
}

func (r Repository) HealthCheck() error {
	return nil
}

func (r Repository) Shutdown() error {
	return r.kv.Shutdown()
}
