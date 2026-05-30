package analyzer

import (
	"context"
	"encoding/json"
	"log/slog"

	"skylytics/internal/core"

	apibsky "github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/jetstream/pkg/models"
)

const feedPostCollection = "app.bsky.feed.post"

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
		if collection == feedPostCollection && operation == models.CommitOperationCreate {
			a.analyzeFeedPost(ctx, event.Commit.Record)
		}
	case models.EventKindAccount:
	case models.EventKindIdentity:
	}

	a.Metrics.IncJetstreamProcessedEventsTotal(ctx, kind, operation, collection)

	return nil
}

func (a *EventAnalyzer) analyzeFeedPost(_ context.Context, record []byte) {
	var post apibsky.FeedPost
	if err := json.Unmarshal(record, &post); err != nil {
		a.Logger.Error("error unmarshalling feed post", "error", err, "record", string(record))
	}
}
