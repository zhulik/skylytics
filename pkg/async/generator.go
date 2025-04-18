package async

import (
	"context"
	"github.com/samber/lo"
)

type Yielder[T any] func(T, error)
type Gen[T any] func(context.Context, Yielder[T]) error

func Generator[T any](ctx context.Context, gen Gen[T]) <-chan Result[T] {
	ch := make(chan Result[T], 1)

	y := func(t T, err error) {
		ch <- NewResult(t, err)
	}

	go func() {
		err := gen(ctx, y)
		if err != nil {
			ch <- NewResult[T](lo.Empty[T](), err)
		}
		close(ch)
	}()

	return ch
}
