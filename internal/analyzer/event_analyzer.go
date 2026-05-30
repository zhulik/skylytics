package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
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
		if collection == feedPostCollection {
			var post apibsky.FeedPost
			if err := json.Unmarshal(event.Commit.Record, &post); err != nil {
				return fmt.Errorf("unmarshal feed post: %w", err)
			}
			a.Logger.Info("analyzing event", "record", post)
		}
	case models.EventKindAccount:
	case models.EventKindIdentity:
	}

	a.Metrics.IncJetstreamProcessedEventsTotal(ctx, kind, operation, collection)

	return nil
}
