package analyzer

import (
	"context"
	"log/slog"
	"skylytics/internal/core"
	"skylytics/internal/leaderboard"
	"time"

	apibsky "github.com/bluesky-social/indigo/api/bsky"
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
	return s.zincrExpire(ctx, leaderboard.FineMinuteKey(interaction, t), postID, 48*time.Hour)
}

func (s *LeaderboardBucketSaver) zincrReplyCounters(ctx context.Context, timestamp, parentID, rootID string) error {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return err
	}
	key := leaderboard.FineMinuteKey(core.Reply, t)
	err = s.zincrExpire(ctx, key, parentID, 48*time.Hour)
	if err != nil {
		return err
	}
	return s.zincrExpire(ctx, key, rootID, 48*time.Hour)
}

func (s *LeaderboardBucketSaver) zincrExpire(ctx context.Context, key, postID string, duration time.Duration) error {
	err := s.Redis.ZIncrBy(ctx, key, 1, postID).Err()
	if err != nil {
		return err
	}
	return s.Redis.Expire(ctx, key, duration).Err()
}
