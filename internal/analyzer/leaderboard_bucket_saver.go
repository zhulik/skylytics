package analyzer

import (
	"context"
	"log/slog"
	"skylytics/internal/core"
	"skylytics/internal/leaderboard"
	"time"

	apibsky "github.com/bluesky-social/indigo/api/bsky"
	"golang.org/x/sync/errgroup"
)

type LeaderboardBucketSaver struct {
	Logger *slog.Logger
	Redis  core.Redis
}

func (s *LeaderboardBucketSaver) SaveLike(ctx context.Context, like *apibsky.FeedLike) error {
	return s.zincrInteractionCounters(ctx, core.Like, like.CreatedAt, like.Subject.Uri)
}

func (s *LeaderboardBucketSaver) SaveRepost(ctx context.Context, repost *apibsky.FeedRepost) error {
	return s.zincrInteractionCounters(ctx, core.Repost, repost.CreatedAt, repost.Subject.Uri)
}

func (s *LeaderboardBucketSaver) SaveQuote(ctx context.Context, createdAt string, embed *apibsky.EmbedRecord) error {
	return s.zincrInteractionCounters(ctx, core.Quote, createdAt, embed.Record.Uri)
}

func (s *LeaderboardBucketSaver) SaveReply(ctx context.Context, createdAt string, reply *apibsky.FeedPost_ReplyRef) error {
	return s.zincrReplyCounters(ctx, createdAt, reply.Parent.Uri, reply.Root.Uri)
}

func (s *LeaderboardBucketSaver) zincrInteractionCounters(ctx context.Context, interaction core.Interaction, timestamp, postID string) error {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return err
	}
	return s.zincrAllBuckets(ctx, interaction, t, postID)
}

func (s *LeaderboardBucketSaver) zincrReplyCounters(ctx context.Context, timestamp, parentID, rootID string) error {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return err
	}
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return s.zincrAllBuckets(ctx, core.Reply, t, parentID) })
	g.Go(func() error { return s.zincrAllBuckets(ctx, core.Reply, t, rootID) })
	return g.Wait()
}

func (s *LeaderboardBucketSaver) zincrAllBuckets(ctx context.Context, interaction core.Interaction, t time.Time, postID string) error {
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return s.zincrExpire(ctx, leaderboard.FineMinuteKey(interaction, t), postID, 48*time.Hour)
	})
	g.Go(func() error {
		return s.zincrExpire(ctx, leaderboard.HourlyKey(interaction, t), postID, 14*24*time.Hour)
	})
	g.Go(func() error {
		return s.zincrExpire(ctx, leaderboard.DailyKey(interaction, t), postID, 60*24*time.Hour)
	})
	return g.Wait()
}

func (s *LeaderboardBucketSaver) zincrExpire(ctx context.Context, key, postID string, duration time.Duration) error {
	err := s.Redis.ZIncrBy(ctx, key, 1, postID).Err()
	if err != nil {
		return err
	}
	return s.Redis.Expire(ctx, key, duration).Err()
}
