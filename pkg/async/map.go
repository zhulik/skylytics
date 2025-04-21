package async

import "context"

type MapIteratee[T any, R any] func(T) R
type MapIterateeI[T any, R any] func(T, int) R
type MapAsyncIteratee[T any, R any] func(context.Context, T) (R, error)
type MapAsyncIterateeI[T any, R any] func(context.Context, T, int) (R, error)

func Map[T any, R any](collection []T, iteratee MapIteratee[T, R]) []R {
	r, _ := AsyncMapI(context.Background(), collection, func(_ context.Context, t T, _ int) (R, error) {
		return iteratee(t), nil
	})
	return r
}

func MapI[T any, R any](collection []T, iteratee MapIterateeI[T, R]) []R {
	r, _ := AsyncMapI(context.Background(), collection, func(_ context.Context, t T, i int) (R, error) {
		return iteratee(t, i), nil
	})
	return r
}

func AsyncMap[T any, R any](ctx context.Context, collection []T, iteratee MapAsyncIteratee[T, R]) ([]R, error) { //nolint:revive
	return AsyncMapI(ctx, collection, func(ctx context.Context, t T, _ int) (R, error) {
		return iteratee(ctx, t)
	})
}

func AsyncMapI[T any, R any](ctx context.Context, collection []T, iteratee MapAsyncIterateeI[T, R]) ([]R, error) { //nolint:revive
	result := make([]R, len(collection))

	for i, item := range collection {
		r, err := iteratee(ctx, item, i)
		if err != nil {
			return nil, err
		}
		result[i] = r
	}

	return result, nil
}
