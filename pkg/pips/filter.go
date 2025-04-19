package pips

import "context"

type filterStage[T any] struct {
	filter func(context.Context, T) (bool, error)
}

func (s filterStage[I]) Run(ctx context.Context, input <-chan D[any]) <-chan D[any] {
	return Subscriber(ctx, input, func(ctx context.Context, item D[any], out chan<- D[any]) error {
		keep, err := s.filter(ctx, item.Value.(I))
		if err != nil {
			return err
		}
		if !keep {
			return nil
		}

		out <- item
		return nil
	})
}

func Filter[I any](filter func(context.Context, I) (bool, error)) Stage {
	return filterStage[I]{
		filter: filter,
	}
}
