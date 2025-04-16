package async

import "github.com/samber/lo"

type Yielder[T any] func(T)

func Generator[T any](gen func(Yielder[T]) error) <-chan Result[T] {
	ch := make(chan Result[T], 1)

	y := func(t T) {
		ch <- NewResult(t)
	}

	go func() {
		err := gen(y)
		if err != nil {
			ch <- NewResult[T](lo.Empty[T](), err)
		}
		close(ch)
	}()

	return ch
}
