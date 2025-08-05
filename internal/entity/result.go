package entity

type Result[T any] struct {
	Value T
	Err   error
}

func NewResult[T any](t T) Result[T] {
	return Result[T]{
		Value: t,
	}
}

func NewResultWithError[T any](err error) Result[T] {
	var t T
	return Result[T]{
		Value: t,
		Err:   err,
	}
}
