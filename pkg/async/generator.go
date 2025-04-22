package async

import (
	"context"

	"github.com/zhulik/pips"
)

type Yielder[T any] func(T, error)
type Gen[T any] func(context.Context, Yielder[T]) error

func Generator[T any](gen Gen[T]) (*JobHandle[any], <-chan pips.D[T]) {
	ch := make(chan pips.D[T])

	y := func(t T, err error) {
		ch <- pips.NewD(t, err)
	}

	handle := Job(func(ctx context.Context) (any, error) {
		defer close(ch)

		return nil, gen(ctx, y)
	})

	return handle, ch
}
