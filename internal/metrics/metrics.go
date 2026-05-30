package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var jetstreamProcessedEventsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "jetstream_processed_events_total",
		Help: "Total number of processed events",
	},
	[]string{"kind", "operation", "collection"},
)
