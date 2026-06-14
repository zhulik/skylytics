package leaderboard

import (
	"fmt"
	"skylytics/internal/core"
	"time"
)

func InteractionKeyPrefix(interaction core.Interaction) string {
	switch interaction {
	case core.Reply:
		return "replies:"
	case core.Like:
		return "likes:"
	case core.Repost:
		return "reposts:"
	case core.Quote:
		return "quotes:"
	default:
		panic(fmt.Errorf("unknown interaction: %s", interaction.Value))
	}
}

func fiveMinuteBucket(t time.Time) string {
	return t.UTC().Truncate(5 * time.Minute).Format("2006-01-02T15:04")
}

func hourBucket(t time.Time) string {
	return t.UTC().Truncate(time.Hour).Format("2006-01-02T15")
}

func FineMinuteKey(interaction core.Interaction, t time.Time) string {
	return InteractionKeyPrefix(interaction) + fiveMinuteBucket(t)
}

func HourlyKey(interaction core.Interaction, hour time.Time) string {
	return InteractionKeyPrefix(interaction) + hourBucket(hour)
}
