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

	eventProcessingDurationSeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "skylytics_event_processing_duration_seconds",
			Help:    "Time spent processing a single JetStream event in the analyzer",
			Buckets: []float64{0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
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

	leaderboardRawBucketTopScore = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "skylytics_leaderboard_raw_bucket_top_score",
			Help: "Highest single-post interaction count in the penultimate 5-minute raw leaderboard bucket",
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
