package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	jetstreamProcessedEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jetstream_processed_events_total",
			Help: "Total number of processed events",
		},
		[]string{"kind", "operation", "collection"},
	)

	jetstreamSubscriptionErrorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "jetstream_subscription_errors_total",
			Help: "Total number of Jetstream subscription errors that triggered a restart",
		},
	)
)
