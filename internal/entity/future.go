package entity

import (
	"context"
	"sync"
)

type Result[T any] struct {
	Data T
	Err  error
}

func NewResult[T any](t T) Result[T] {
	return Result[T]{Data: t}
}

func NewResultWithError[T any](err error) Result[T] {
	var t T
	return Result[T]{Data: t, Err: err}
}

type Future[T any] struct {
	input        chan Result[T]
	resolved     bool
	value        []Result[T]
	cancel       context.CancelFunc
	lock         sync.Mutex
	stopCallback func(canceled bool)
}

func NewFuture[T any](ctx context.Context, input chan Result[T]) *Future[T] {
	fctx, cancel := context.WithCancel(ctx)

	f := &Future[T]{
		input:    input,
		resolved: false,
		cancel:   cancel,
	}

	go func(ctx context.Context) {
		for value := range f.input {
			f.value = append(f.value, value)
			select {
			case <-ctx.Done():
				// keep draining the input
				go func() {
					for _ = range input {
					}
				}()
				goto outFor
			default:
			}
		}
	outFor:
		f.lock.Lock()
		defer f.lock.Unlock()

		f.resolved = true
		if f.stopCallback != nil {
			f.stopCallback(fctx.Err() != nil)
		}
	}(fctx)

	return f
}

func (f *Future[T]) StopCallback(fn func(canceled bool)) {
	f.stopCallback = fn
}

func (f *Future[T]) IsResolved() bool {
	f.lock.Lock()
	defer f.lock.Unlock()
	return f.resolved
}

func (f *Future[T]) Poll() (value []Result[T], isResolved bool) {
	if f.IsResolved() {
		f.cancel()
		return f.value, true
	}

	return nil, false
}

func (f *Future[T]) Cancel() {
	if !f.IsResolved() {
		f.cancel()
	}
}
