package async

import (
	"context"
	"sync/atomic"
)

func WorkerPool[T any](ctx context.Context, concurrency int, ch <-chan T, fn EachAsyncIteratee[T]) error {
	semaphore := make(chan struct{}, concurrency)
	defer close(semaphore)

	aErr := atomic.Pointer[error]{}

	for m := range ch {
		semaphore <- struct{}{}

		err := aErr.Load()
		if err != nil {
			return *err
		}

		go func(m T) {
			defer func() { <-semaphore }()
			err := fn(ctx, m)

			if err != nil {
				aErr.Store(&err)
			}
		}(m)
	}

	err := aErr.Load()
	if err != nil {
		return *err
	}
	return nil
}
