package analyzer

import (
	"context"
	"encoding/json"
	"log/slog"

	"skylytics/internal/core"

	apibsky "github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/jetstream/pkg/models"
)

const (
	postCollection   = "app.bsky.feed.post"
	likeCollection   = "app.bsky.feed.like"
	repostCollection = "app.bsky.feed.repost"

	interactionLike   = "like"
	interactionRepost = "repost"
	interactionQuote  = "quote"
	interactionReply  = "reply"
)

type EventAnalyzer struct {
	Logger  *slog.Logger
	Metrics core.MetricsCollector

	LeaderboardRawBucketSaver *LeaderboardRawBucketSaver
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
	case likeCollection:
		a.analyzeLikeCreated(ctx, event.Commit.Record)
	case repostCollection:
		a.analyzeRepostCreated(ctx, event.Commit.Record)
	}
}

func (a *EventAnalyzer) analyzePostCreated(ctx context.Context, record []byte) {
	var post apibsky.FeedPost
	if err := json.Unmarshal(record, &post); err != nil {
		a.Logger.Error("error unmarshalling feed post", "error", err, "record", string(record))
	}

	var imagesCount int

	if embed := post.Embed; embed != nil {
		if embedImages := embed.EmbedImages; embedImages != nil {
			imagesCount = len(embedImages.Images)
		}

		if embedRecord := embed.EmbedRecord; embedRecord != nil {
			err := a.LeaderboardRawBucketSaver.SaveQuote(ctx, post.CreatedAt, embedRecord)
			if err != nil {
				a.Logger.Error("error saving quote", "error", err)
			}
			a.Metrics.IncPostInteractionsTotal(ctx, interactionQuote)
		}

		if embedRecordWithMedia := embed.EmbedRecordWithMedia; embedRecordWithMedia != nil {
			err := a.LeaderboardRawBucketSaver.SaveQuote(ctx, post.CreatedAt, embedRecordWithMedia.Record)
			if err != nil {
				a.Logger.Error("error saving quote", "error", err)
			}
			a.Metrics.IncPostInteractionsTotal(ctx, interactionQuote)
		}
	}

	if post.Reply != nil {
		err := a.LeaderboardRawBucketSaver.SaveReply(ctx, post.CreatedAt, post.Reply)
		if err != nil {
			a.Logger.Error("error saving reply", "error", err)
		}
		a.Metrics.IncPostInteractionsTotal(ctx, interactionReply)
	}

	a.Metrics.IncBlueskyPostsTotal(ctx, len(post.Langs), imagesCount)

	for _, lang := range post.Langs {
		a.Metrics.IncBlueskyPostsByLanguageTotal(ctx, lang)
	}
}

func (a *EventAnalyzer) analyzeLikeCreated(ctx context.Context, record []byte) {
	var like apibsky.FeedLike
	if err := json.Unmarshal(record, &like); err != nil {
		a.Logger.Error("error unmarshalling feed like", "error", err, "record", string(record))
	}
	err := a.LeaderboardRawBucketSaver.SaveLike(ctx, &like)
	if err != nil {
		a.Logger.Error("error saving like", "error", err)
	}
	a.Metrics.IncPostInteractionsTotal(ctx, interactionLike)
}

func (a *EventAnalyzer) analyzeRepostCreated(ctx context.Context, record []byte) {
	var repost apibsky.FeedRepost
	if err := json.Unmarshal(record, &repost); err != nil {
		a.Logger.Error("error unmarshalling feed repost", "error", err, "record", string(record))
	}
	err := a.LeaderboardRawBucketSaver.SaveRepost(ctx, &repost)
	if err != nil {
		a.Logger.Error("error saving repost", "error", err)
	}
	a.Metrics.IncPostInteractionsTotal(ctx, interactionRepost)
}
