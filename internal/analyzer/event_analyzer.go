package analyzer

import (
	"context"
	"encoding/json"
	"log/slog"

	"skylytics/internal/core"

	apibsky "github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/jetstream/pkg/models"
)

const postCollection = "app.bsky.feed.post"

type EventAnalyzer struct {
	Logger  *slog.Logger
	Metrics core.MetricsCollector
}

func (a *EventAnalyzer) Analyze(ctx context.Context, event *models.Event) error {
	kind := event.Kind
	operation := ""
	collection := ""

	switch kind {
	case models.EventKindCommit:
		operation = event.Commit.Operation
		collection = event.Commit.Collection

		if operation == models.CommitOperationCreate {
			a.analyzeCreatedCommit(ctx, event)
		}
	case models.EventKindAccount:
	case models.EventKindIdentity:
	}

	a.Metrics.IncJetstreamProcessedEventsTotal(ctx, kind, operation, collection)

	return nil
}

func (a *EventAnalyzer) analyzeCreatedCommit(ctx context.Context, event *models.Event) {
	switch event.Commit.Collection {
	case postCollection:
		a.analyzePostCreated(ctx, event.Commit.Record)
	}
}

func (a *EventAnalyzer) analyzePostCreated(_ context.Context, record []byte) {
	var post apibsky.FeedPost
	if err := json.Unmarshal(record, &post); err != nil {
		a.Logger.Error("error unmarshalling feed post", "error", err, "record", string(record))
	}
}
