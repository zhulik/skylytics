package async

import (
	"context"
	"sync/atomic"

	"github.com/zhulik/pips"
)

type JobHandle[T any] struct {
	cancel func()
	done   chan pips.D[T]
	err    atomic.Pointer[error]
}

func Job[T any](job func(ctx context.Context) (T, error)) *JobHandle[T] {
	ctx, cancel := context.WithCancel(context.Background())
	handle := JobHandle[T]{
		cancel: cancel,
		done:   make(chan pips.D[T], 1),
	}

	go func() {
		defer cancel()

		res, err := job(ctx)

		handle.err.Store(&err)
		handle.done <- pips.NewD(res, err)
	}()

	return &handle
}

func (j *JobHandle[T]) Stop() {
	j.cancel()
}

func (j *JobHandle[T]) StopWait() (T, error) {
	j.Stop()

	return j.Wait()
}

func (j *JobHandle[T]) Wait() (T, error) {
	return (<-j.done).Unpack()
}

func (j *JobHandle[T]) Error() error {
	var err = j.err.Load()
	if err == nil {
		return nil
	}
	return *err
}
