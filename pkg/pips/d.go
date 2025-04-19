package pips

import (
	"github.com/samber/lo"
)

type D[T any] struct {
	Value T
	Err   error
}

func AnyD[T any](value T) D[any] {
	return NewD[any](value)
}

func NewD[T any](value T) D[T] {
	return D[T]{value, nil}
}

func ErrD[T any](err error) D[T] {
	return D[T]{lo.Empty[T](), err}
}

func CastD[I any, O any](d D[I]) D[O] {
	if d.Err != nil {
		return ErrD[O](d.Err)
	}
	return NewD(any(d.Value).(O))
}

func (r D[T]) Unpack() (T, error) {
	return r.Value, r.Err
}

func (r D[T]) ToAny() D[any] {
	return CastD[T, any](r)
}
