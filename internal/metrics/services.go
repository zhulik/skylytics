package metrics

import (
	"skylytics/internal/core"

	"github.com/zhulik/pal"
)

func Provide() pal.ServiceDef {
	return pal.ProvideList(
		pal.Provide[core.MetricsCollector](&Collector{}),
	)
}
