package metrics

import (
	"skylytics/internal/core"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	JetstreamProcessedEventsTotal core.MetricName = "jetstream_processed_events_total"
)

var (
	counters = map[core.MetricName]*prometheus.CounterVec{
		JetstreamProcessedEventsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: string(JetstreamProcessedEventsTotal),
				Help: "Total number of processed events",
			},
			[]string{"kind", "operation", "collection"},
		),
	}
)
