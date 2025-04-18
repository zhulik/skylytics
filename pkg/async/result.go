package async

import (
	"context"
)

type Result[T any] struct {
	Value T
	Err   error
}

func NewResult[T any](value T, errs ...error) Result[T] {
	var err error
	if len(errs) > 0 {
		err = errs[0]
	}
	return Result[T]{Value: value, Err: err}
}

func (r Result[T]) Unpack() (T, error) {
	return r.Value, r.Err
}

func UnpackAll[T any](results []Result[T]) ([]T, error) {
	return AsyncMap(nil, results, func(ctx context.Context, result Result[T]) (T, error) {
		return result.Unpack()
	})
}
