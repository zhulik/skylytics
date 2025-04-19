package apply

import (
	"context"
	"skylytics/pkg/pips"
)

type pipelineStage[I any, O any] struct {
	pipeline *pips.Pipeline[I, O]
}

func (p pipelineStage[I, O]) Run(ctx context.Context, input <-chan pips.D[any]) <-chan pips.D[any] {
	return pips.CastDChan[O, any](ctx, p.pipeline.Run(ctx, pips.MapChan(ctx, input, func(d pips.D[any]) I {
		return d.Value.(I)
	})))
}

func Pipeline[I any, O any](pipeline *pips.Pipeline[I, O]) pips.Stage {
	return pipelineStage[I, O]{pipeline}
}
