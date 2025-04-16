package async

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
