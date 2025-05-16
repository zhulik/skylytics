package events

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"skylytics/internal/core"
)

const (
	// KVBucketName is the name of the KV bucket for events
	KVBucketName = "skylytics-events"
)

type Repository struct {
	JS core.JetstreamClient

	KV core.KeyValueClient
}

func (r *Repository) Init(ctx context.Context) error {
	kv, err := r.JS.KV(ctx, KVBucketName)
	if err != nil {
		return fmt.Errorf("failed to create KV client: %w", err)
	}
	r.KV = kv

	return nil
}

func (r *Repository) Insert(ctx context.Context, events ...core.EventModel) error {
	for i, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event %d: %w", i, err)
		}

		if err := r.KV.Put(ctx, key(event), data); err != nil {
			return fmt.Errorf("failed to store event %d, %w", i, err)
		}
	}

	return nil
}

func key(event core.EventModel) string {
	hashedDID := sha256.Sum256([]byte(event.Event.Did))
	return fmt.Sprintf("event.%x.%d", hashedDID, event.Event.TimeUS)
}
