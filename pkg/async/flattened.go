package async

import (
	"context"
	"github.com/samber/lo"
)

func Flattened[T any](ctx context.Context, ch <-chan Result[[]T]) <-chan Result[T] {
	flatChan := make(chan Result[T], 1)

	go func() {
		for m := range ch {
			items, err := m.Unpack()
			if err != nil {
				flatChan <- NewResult(lo.Empty[T](), err)
				return
			}

			for _, item := range items {
				select {
				case <-ctx.Done():
					return
				default:
					flatChan <- NewResult(item)
				}
			}
		}
	}()

	return flatChan
}
