package async

import (
	"context"

	"github.com/samber/lo"

	"github.com/zhulik/pips"
)

type Yielder[T any] func(T, error)
type Gen[T any] func(context.Context, Yielder[T]) error

func Generator[T any](ctx context.Context, gen Gen[T]) <-chan pips.D[T] {
	ch := make(chan pips.D[T])

	y := func(t T, err error) {
		ch <- pips.NewD(t, err)
	}

	go func() {
		err := gen(ctx, y)
		if err != nil {
			ch <- pips.NewD[T](lo.Empty[T](), err)
		}
		close(ch)
	}()

	return ch
}
