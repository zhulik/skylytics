package async

import (
	"context"
)

type JobHandle[T any] struct {
	cancel func()
	done   chan Result[T]
}

func Job[T any](job func(ctx context.Context) (T, error)) *JobHandle[T] {
	ctx, cancel := context.WithCancel(context.Background())
	handle := JobHandle[T]{
		cancel: cancel,
		done:   make(chan Result[T], 1),
	}

	go func() {
		res, err := job(ctx)

		handle.done <- NewResult(res, err)

		cancel()
	}()

	return &handle
}

func (j JobHandle[T]) Stop() {
	j.cancel()
}

func (j JobHandle[T]) Wait() (T, error) {
	return (<-j.done).Unpack()
}
