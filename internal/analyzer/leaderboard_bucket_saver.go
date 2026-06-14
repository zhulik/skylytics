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
	createdAt, err := time.Parse(time.RFC3339, like.CreatedAt)
	if err != nil {
		return err
	}
	return s.ZincrExpire(ctx, leaderboard.FineMinuteKey(core.Like, createdAt), like.Subject.Uri)
}

func (s *LeaderboardBucketSaver) SaveRepost(ctx context.Context, repost *apibsky.FeedRepost) error {
	createdAt, err := time.Parse(time.RFC3339, repost.CreatedAt)
	if err != nil {
		return err
	}
	return s.ZincrExpire(ctx, leaderboard.FineMinuteKey(core.Repost, createdAt), repost.Subject.Uri)
}

func (s *LeaderboardBucketSaver) SaveQuote(ctx context.Context, createdAt string, embed *apibsky.EmbedRecord) error {
	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return err
	}
	return s.ZincrExpire(ctx, leaderboard.FineMinuteKey(core.Quote, t), embed.Record.Uri)
}

func (s *LeaderboardBucketSaver) SaveReply(ctx context.Context, createdAt string, reply *apibsky.FeedPost_ReplyRef) error {
	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return err
	}
	key := leaderboard.FineMinuteKey(core.Reply, t)
	err = s.ZincrExpire(ctx, key, reply.Parent.Uri)
	if err != nil {
		return err
	}
	return s.ZincrExpire(ctx, key, reply.Root.Uri)
}

func (s *LeaderboardBucketSaver) ZincrExpire(ctx context.Context, key, postID string) error {
	err := s.Redis.ZIncrBy(ctx, key, 1, postID).Err()
	if err != nil {
		return err
	}
	return s.Redis.Expire(ctx, key, 48*time.Hour).Err()
}
