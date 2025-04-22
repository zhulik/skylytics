package async

import (
	"context"

	"github.com/samber/lo"

	"github.com/zhulik/pips"
)

type Yielder[T any] func(T, error)
type Gen[T any] func(context.Context, Yielder[T]) error

func Generator[T any](ctx context.Context, gen Gen[T]) (JobHandle[T], <-chan pips.D[T]) {
	ch := make(chan pips.D[T])

	y := func(t T, err error) {
		ch <- pips.NewD(t, err)
	}

	handle := async.Job(func(ctx context.Context) (any, error) {
		defer close(ch)
		err := gen(ctx, y)
		return nil, err
	})

	return handle, ch
}
