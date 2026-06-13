package leaderboard

import (
	"time"

	"github.com/orsinium-labs/enum"
)

type Interaction enum.Member[string]

var (
	Like            = Interaction{"like"}
	Repost          = Interaction{"repost"}
	Reply           = Interaction{"reply"}
	Quote           = Interaction{"quote"}
	AllInteractions = enum.New(Like, Repost, Reply, Quote)
)

func RawKeyPrefix(interaction Interaction) string {
	switch interaction {
	case Reply:
		return "replies:"
	case Like:
		return "likes:"
	case Repost:
		return "reposts:"
	case Quote:
		return "quotes:"
	default:
		panic("unknown interaction: " + interaction.Value)
	}
}

func HourlyKeyPrefix(interaction Interaction) string {
	switch interaction {
	case Like:
		return "hourly_likes:"
	case Repost:
		return "hourly_reposts:"
	case Reply:
		return "hourly_replies:"
	case Quote:
		return "hourly_quotes:"
	default:
		panic("unknown interaction: " + interaction.Value)
	}
}

func fiveMinuteBucket(t time.Time) string {
	return t.UTC().Truncate(5 * time.Minute).Format("2006-01-02T15:04")
}

func HourBucket(t time.Time) string {
	return t.UTC().Truncate(time.Hour).Format("2006-01-02T15")
}

func RawKey(interaction Interaction, t time.Time) string {
	return RawKeyPrefix(interaction) + fiveMinuteBucket(t)
}

func HourlyKey(interaction Interaction, hour time.Time) string {
	return HourlyKeyPrefix(interaction) + HourBucket(hour)
}

func LastCompleteHour(now time.Time) time.Time {
	penultimate := now.UTC().Truncate(5 * time.Minute).Add(-5 * time.Minute)
	hour := penultimate.Truncate(time.Hour)
	if penultimate.Sub(hour) < 55*time.Minute {
		hour = hour.Add(-time.Hour)
	}
	return hour
}

func RawKeysForHour(interaction Interaction, hour time.Time) []string {
	keys := make([]string, 12)
	for i := range 12 {
		keys[i] = RawKey(interaction, hour.Add(time.Duration(i*5)*time.Minute))
	}
	return keys
}
