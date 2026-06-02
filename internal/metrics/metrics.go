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

	jetstreamSubscriptionErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "skylytics_jetstream_subscription_errors_total",
			Help: "Total number of Jetstream subscription errors that triggered a restart",
		},
		[]string{"error"},
	)

	blueskyPostCreatedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "skylytics_bluesky_post_created",
			Help: "Total number of Bluesky posts created",
		},
		[]string{"language_count", "image_count"},
	)

	blueskyPostCreatedInLanguageTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "skylytics_bluesky_post_created_in_language_total",
			Help: "Total number of Bluesky posts created, counted once per language tag on the post",
		},
		[]string{"language"},
	)

	rawBucketsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "skylytics_raw_buckets_total",
			Help: "Total number of raw leaderboard buckets in Redis",
		},
		[]string{"content"},
	)

	postInteractedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "skylytics_post_interacted",
			Help: "Total number of post interactions (likes, reposts, quotes, replies)",
		},
		[]string{"interation"},
	)
)
