package apply

import (
	"context"
	"skylytics/pkg/pips"
)

type mapStage[I any, O any] struct {
	mapper func(context.Context, I) (O, error)
}

func (m mapStage[I, O]) Run(ctx context.Context, input <-chan pips.D[any]) <-chan pips.D[any] {
	return Subscriber(ctx, input, func(ctx context.Context, item pips.D[any], out chan<- pips.D[any]) error {
		res, err := m.mapper(ctx, item.Value.(I))
		if err != nil {
			return err
		}

		out <- pips.AnyD(res)

		return nil
	})
}

func Map[I any, O any](mapper func(context.Context, I) (O, error)) pips.Stage {
	return mapStage[I, O]{mapper}
}
