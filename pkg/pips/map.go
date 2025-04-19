package pips

import "context"

type mapStage[I any, O any] struct {
	mapper func(context.Context, I) (O, error)
}

func (m mapStage[I, O]) Run(ctx context.Context, input <-chan D[any]) <-chan D[any] {
	return Subscriber(ctx, input, func(ctx context.Context, item D[any], out chan<- D[any]) error {
		res, err := m.mapper(ctx, item.Value.(I))
		if err != nil {
			return err
		}

		out <- NewD[any](res)

		return nil
	})
}

func Map[I any, O any](mapper func(context.Context, I) (O, error)) Stage {
	return mapStage[I, O]{
		mapper: mapper,
	}
}
