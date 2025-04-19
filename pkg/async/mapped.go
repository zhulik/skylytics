package async

import (
	"context"
	"github.com/samber/lo"
)

func Mapped[T any, R any](ctx context.Context, ch <-chan Result[T], iteratee MapAsyncIteratee[T, R]) <-chan Result[R] {
	flatChan := make(chan Result[R], 1)

	go func() {
		for m := range ch {
			select {
			case <-ctx.Done():

				return
			default:
				item, err := m.Unpack()
				if err != nil {
					flatChan <- NewResult(lo.Empty[R](), err)
					return
				}

				r, err := iteratee(ctx, item)
				flatChan <- NewResult(r, err)
				if err != nil {
					return
				}
			}
		}
	}()

	return flatChan
}
