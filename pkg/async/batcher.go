package async

import (
	"context"
	"time"
)

func Batcher[T any](ctx context.Context, ch <-chan T, batchSize int, timeout time.Duration) <-chan []T {
	batchChan := make(chan []T, 1)

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

			batchChan <- buffer
			buffer = make([]T, 0, batchSize)
		}

		for {
			select {
			case <-ctx.Done():
				sendReset()
				return

			case <-ticker.C:
				sendReset()

			case item, ok := <-ch:
				if !ok {
					sendReset()
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
