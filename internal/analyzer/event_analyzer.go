package analyzer

import (
	"context"

	"skylytics/internal/core"

	"github.com/bluesky-social/jetstream/pkg/models"
)

type EventAnalyzer struct {
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
	case models.EventKindAccount:
	case models.EventKindIdentity:
	}

	a.Metrics.IncJetstreamProcessedEventsTotal(ctx, kind, operation, collection)

	return nil
}
