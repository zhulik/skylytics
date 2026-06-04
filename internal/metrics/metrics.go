package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	jetstreamProcessedEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "skylytics_jetstream_processed_events_total",
			Help: "Total number of processed events",
		},
		[]string{"kind", "operation", "collection"},
	)

	jetstreamEventLagSeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "skylytics_jetstream_event_lag_seconds",
			Help:    "Lag between JetStream event time_us and when the event is processed",
			Buckets: []float64{0.5, 1, 2, 5, 10, 30, 60, 120, 300, 600, 1800},
		},
	)

	jetstreamSubscriptionErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "skylytics_jetstream_subscription_errors_total",
			Help: "Total number of Jetstream subscription errors that triggered a restart",
		},
		[]string{"error"},
	)

	blueskyPostsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "skylytics_bluesky_posts_total",
			Help: "Total number of Bluesky posts created",
		},
		[]string{"language_count", "image_count"},
	)

	blueskyPostsByLanguageTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "skylytics_bluesky_posts_by_language_total",
			Help: "Total number of Bluesky posts created, counted once per language tag on the post",
		},
		[]string{"language"},
	)

	leaderboardRawBucketKeysTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "skylytics_leaderboard_raw_bucket_keys",
			Help: "Total number of raw leaderboard buckets in Redis",
		},
		[]string{"content"},
	)

	leaderboardRawBucketMembersTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "skylytics_leaderboard_raw_bucket_members",
			Help: "Total distinct post URIs stored across all raw leaderboard buckets (sum of ZCARD per bucket)",
		},
		[]string{"content"},
	)

	postInteractionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "skylytics_post_interactions_total",
			Help: "Total number of post interactions (likes, reposts, quotes, replies)",
		},
		[]string{"interaction"},
	)
)
