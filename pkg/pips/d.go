package pips

import (
	"errors"
	"github.com/samber/lo"
)

type D[T any] struct {
	Value T
	Err   error
}

func NewD[T any](value T, errs ...error) D[T] {
	return D[T]{Value: value, Err: errors.Join(errs...)}
}

func ErrD[T any](err error) D[T] {
	return D[T]{Value: lo.Empty[T](), Err: err}
}

func (r D[T]) Unpack() (T, error) {
	return r.Value, r.Err
}

func (r D[T]) ToAny() D[any] {
	return D[any]{Value: r.Value, Err: r.Err}
}
