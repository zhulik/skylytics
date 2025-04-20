package apply

import (
	"context"
	"skylytics/pkg/pips"
)

type batchStage struct {
	size int
}

func (s batchStage) Run(ctx context.Context, input <-chan pips.D[any]) <-chan pips.D[any] {
	batchChan := make(chan pips.D[any])

	buffer := make([]any, 0, s.size)

	sendReset := func() {
		if len(buffer) == 0 {
			return
		}

		batchChan <- pips.AnyD(buffer)
		buffer = make([]any, 0, s.size)
	}

	go func() {
		defer close(batchChan)

		for {
			select {
			case <-ctx.Done():
				sendReset()
				return

			case res, ok := <-input:
				if !ok {
					sendReset()
					return
				}
				item, err := res.Unpack()
				if err != nil {
					batchChan <- pips.ErrD[any](err)
					return
				}
				buffer = append(buffer, item)

				if len(buffer) >= s.size {
					sendReset()
				}
			}
		}
	}()

	return batchChan
}

func Batch(batchSize int) pips.Stage {
	return batchStage{
		size: batchSize,
	}
}
