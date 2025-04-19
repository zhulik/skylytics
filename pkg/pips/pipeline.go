package pips

import (
	"context"
)

type Pipeline[I any, O any] struct {
	stages stages[I, O]
}

func New[I any, O any]() *Pipeline[I, O] {
	return &Pipeline[I, O]{
		stages: stages[I, O]{},
	}
}

func (p *Pipeline[I, O]) Then(stage Stage) *Pipeline[I, O] {
	p.stages = append(p.stages, stage)
	return p
}

func (p *Pipeline[I, O]) Run(ctx context.Context, input <-chan D[I]) <-chan D[O] {
	inChan := make(chan D[any])

	var prevOut <-chan D[any]

	prevOut = inChan

	for _, stage := range p.stages {
		prevOut = stage.Run(ctx, prevOut)
	}

	outCh := CastDChan[any, O](ctx, prevOut)

	go func() {
		defer close(inChan)

		for {
			select {
			case <-ctx.Done():
				return
			case in, ok := <-input:
				if !ok {
					return
				}
				inChan <- in.ToAny()
			}
		}
	}()

	return outCh
}

func (p *Pipeline[I, O]) ToStage() Stage {
	return pipelineStage[I, O]{p}
}

type pipelineStage[I any, O any] struct {
	p *Pipeline[I, O]
}

func (p pipelineStage[I, O]) Run(ctx context.Context, input <-chan D[any]) <-chan D[any] {
	return CastDChan[O, any](ctx, p.p.Run(ctx, CastDChan[any, I](ctx, input)))
}
