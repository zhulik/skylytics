package analyzer

import (
	"skylytics/internal/core"

	"github.com/zhulik/pal"
)

func Provide() pal.ServiceDef {
	return pal.Provide[core.EventAnalyzer](&EventAnalyzer{})
}
