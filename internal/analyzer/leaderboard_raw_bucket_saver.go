package analyzer

import (
	"context"
	"log/slog"
	"skylytics/internal/core"
	"time"

	apibsky "github.com/bluesky-social/indigo/api/bsky"
)

type LeaderboardRawBucketSaver struct {
	Logger *slog.Logger
	Redis  core.Redis
}

func (s *LeaderboardRawBucketSaver) SaveLike(ctx context.Context, like *apibsky.FeedLike) error {
	createdAt, err := time.Parse(time.RFC3339, like.CreatedAt)
	if err != nil {
		return err
	}
	return s.ZincrExpire(ctx, likesKey(createdAt), like.Subject.Uri)
}

func (s *LeaderboardRawBucketSaver) SaveRepost(ctx context.Context, repost *apibsky.FeedRepost) error {
	createdAt, err := time.Parse(time.RFC3339, repost.CreatedAt)
	if err != nil {
		return err
	}
	return s.ZincrExpire(ctx, repostsKey(createdAt), repost.Subject.Uri)
}

func (s *LeaderboardRawBucketSaver) SaveQuote(ctx context.Context, createdAt string, embed *apibsky.EmbedRecord) error {
	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return err
	}
	return s.ZincrExpire(ctx, quotesKey(t), embed.Record.Uri)
}

func (s *LeaderboardRawBucketSaver) ZincrExpire(ctx context.Context, key, postID string) error {
	err := s.Redis.ZIncrBy(ctx, key, 1, postID).Err()
	if err != nil {
		return err
	}
	return s.Redis.Expire(ctx, key, 48*time.Hour).Err()
}

func fiveMinuteBucket(t time.Time) string {
	return t.UTC().Truncate(5 * time.Minute).Format("2006-01-02T15:04")
}

func likesKey(t time.Time) string {
	return "likes:" + fiveMinuteBucket(t)
}

func repostsKey(t time.Time) string {
	return "reposts:" + fiveMinuteBucket(t)
}

func quotesKey(t time.Time) string {
	return "quotes:" + fiveMinuteBucket(t)
}
