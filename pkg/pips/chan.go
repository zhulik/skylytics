package pips

import (
	"context"
)

func ChanInput[T any](ctx context.Context, ch <-chan T) <-chan D[T] {
	return MapChan(ctx, ch, func(d T) D[T] {
		return NewD(d)
	})
}

func CastDChan[I any, O any](ctx context.Context, ch <-chan D[I]) <-chan D[O] {
	return MapChan(ctx, ch, func(d D[I]) D[O] {
		if d.Err != nil {
			return ErrD[O](d.Err)
		}

		return NewD(any(d.Value).(O))
	})
}

func MapChan[I any, O any](ctx context.Context, input <-chan I, f func(I) O) <-chan O {
	out := make(chan O)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return

			case res, ok := <-input:
				if !ok {
					return
				}
				out <- f(res)
			}
		}
	}()

	return out
}
