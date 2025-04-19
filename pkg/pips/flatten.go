package pips

import (
	"context"
)

type flattenStage struct {
}

func (f flattenStage) Run(ctx context.Context, input <-chan D[any]) <-chan D[any] {
	return Subscriber(ctx, input, func(ctx context.Context, item D[any], out chan<- D[any]) error {
		ary, err := item.Unpack()
		if err != nil {
			return err
		}
		for _, item := range ary.([]any) {
			out <- NewD[any](item)
		}
		return nil
	})
}

func Flatten() Stage {
	return flattenStage{}
}
