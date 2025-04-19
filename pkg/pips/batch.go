package pips

import (
	"context"
)

type batchStage struct {
	size int
}

func (s batchStage) Run(ctx context.Context, input <-chan D[any]) <-chan D[any] {
	batchChan := make(chan D[any])

	buffer := make([]any, 0, s.size)

	sendReset := func() {
		if len(buffer) == 0 {
			return
		}

		batchChan <- NewD[any](buffer)
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
					batchChan <- ErrD[any](err)
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

func Batch(batchSize int) Stage {
	return batchStage{}
}
