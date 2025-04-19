package async

import (
	"context"
	"github.com/samber/lo"
	"time"
)

func Batched[T any](ctx context.Context, ch <-chan Result[T], batchSize int, timeout time.Duration) <-chan Result[[]T] {
	batchChan := make(chan Result[[]T], 1)

	go func() {
		ticker := time.NewTicker(timeout)

		defer func() {
			ticker.Stop()
			close(batchChan)
		}()

		buffer := make([]T, 0, batchSize)

		sendReset := func() {
			if len(buffer) == 0 {
				return
			}

			batchChan <- NewResult(buffer)
			buffer = make([]T, 0, batchSize)
		}

		for {
			select {
			case <-ctx.Done():
				sendReset()
				return

			case <-ticker.C:
				sendReset()

			case res, ok := <-ch:
				if !ok {
					sendReset()
					return
				}
				item, err := res.Unpack()
				if err != nil {
					batchChan <- NewResult(lo.Empty[[]T](), err)
					return
				}
				buffer = append(buffer, item)

				ticker.Reset(timeout)

				if len(buffer) >= batchSize {
					sendReset()
				}
			}
		}
	}()

	return batchChan
}
