package pips

import (
	"context"
)

type Stage interface {
	Run(context.Context, <-chan D[any]) <-chan D[any]
}

type Pipeline[I any, O any] struct {
	stages []Stage
}

func New[I any, O any](stages ...Stage) *Pipeline[I, O] {
	return &Pipeline[I, O]{stages}
}

func (p *Pipeline[I, O]) Then(stages ...Stage) *Pipeline[I, O] {
	p.stages = append(p.stages, stages...)
	return p
}

func (p *Pipeline[I, O]) Run(ctx context.Context, input <-chan I) <-chan D[O] {
	inChan := make(chan D[any])

	var prevOut <-chan D[any]

	prevOut = inChan

	for _, stage := range p.stages {
		prevOut = stage.Run(ctx, prevOut)
	}

	outCh := CastDChan[any, O](ctx, prevOut)

	go func() {
		MapToChan(ctx, input, inChan, AnyD)
		defer close(inChan)
	}()

	return outCh
}
