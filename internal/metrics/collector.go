package metrics

import (
	"github.com/samber/do"
	"skylytics/internal/core"
)

type Collector struct{}

func NewCollector(i *do.Injector) (core.MetricsCollector, error) {
	return &Collector{}, nil
}
