package async

import (
	"context"
)

type EachIteratee[T any] func(T)
type EachIterateeI[T any] func(T, int)
type EachAsyncIteratee[T any] func(context.Context, T) error
type EachAsyncIterateeI[T any] func(context.Context, T, int) error

func Each[T any](collection []T, iteratee EachIteratee[T]) {
	AsyncMapI(nil, collection, func(_ context.Context, t T, _ int) (any, error) {
		iteratee(t)
		return nil, nil
	})
}

func EachI[T any](collection []T, iteratee EachIterateeI[T]) {
	AsyncMapI(nil, collection, func(_ context.Context, t T, i int) (any, error) {
		iteratee(t, i)
		return nil, nil
	})
}

func AsyncEach[T any](ctx context.Context, collection []T, iteratee EachAsyncIteratee[T]) error {
	_, err := AsyncMapI(ctx, collection, func(ctx context.Context, t T, _ int) (any, error) {
		return nil, iteratee(ctx, t)
	})
	return err
}

func AsyncEachI[T any](ctx context.Context, collection []T, iteratee EachAsyncIterateeI[T]) error {
	_, err := AsyncMapI(ctx, collection, func(_ context.Context, t T, i int) (any, error) {
		return nil, iteratee(ctx, t, i)
	})
	return err
}
