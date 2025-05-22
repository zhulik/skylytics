package async

import (
	"context"
	"reflect"
	"sync/atomic"

	"github.com/samber/lo"

	"github.com/zhulik/pips"
)

type JobHandle[T any] struct {
	cancel func()
	done   chan pips.D[T]
	result atomic.Pointer[pips.D[T]]
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
		d := pips.NewD(res, err)
		handle.result.Store(&d)
		handle.done <- d
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
	p := j.result.Load()
	if p == nil {
		return (<-j.done).Unpack()
	}

	return (*p).Unpack()
}

func (j *JobHandle[T]) Error() error {
	p := j.result.Load()
	if p == nil {
		return nil
	}

	return (*p).Error()
}

func FirstFailed[T any](ctx context.Context, handles ...*JobHandle[T]) error {
	cases := lo.Map(handles, func(item *JobHandle[T], _ int) reflect.SelectCase {
		return reflect.SelectCase{
			Chan: reflect.ValueOf(item.done),
			Dir:  reflect.SelectRecv,
		}
	})

	cases = append(cases, reflect.SelectCase{
		Chan: reflect.ValueOf(ctx.Done()),
		Dir:  reflect.SelectRecv,
	})
	for {
		chosen, _, ok := reflect.Select(cases)
		if !ok {
			return nil
		}
		if chosen == len(cases)-1 {
			// context is canceled
			return nil
		}

		err := handles[chosen].Error()
		if err != nil {
			return err
		}
	}
}
