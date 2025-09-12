/*
*
Based on https://research.swtch.com/coro
*
*/
package entity

import (
	"errors"
	"fmt"
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

var ErrCanceled = errors.New("coroutine canceled")

type msg[T any] struct {
	panic any
	val   T
}

func newCoroutine[In, Out any](f func(in In, yield func(Out) In) Out) (resume func(In) (Out, bool), cancel func()) {
	cin := make(chan msg[In])
	cout := make(chan msg[Out])
	running := true
	resume = func(in In) (out Out, ok bool) {
		if !running {
			return
		}
		cin <- msg[In]{val: in}
		m := <-cout
		if m.panic != nil {
			panic(m.panic)
		}
		return m.val, running
	}
	cancel = func() {
		if !running {
			return
		}
		e := fmt.Errorf("%w", ErrCanceled) // unique wrapper
		cin <- msg[In]{panic: e}
		m := <-cout
		if m.panic != nil && m.panic != e {
			panic(m.panic)
		}
	}
	yield := func(out Out) In {
		cout <- msg[Out]{val: out}
		m := <-cin
		if m.panic != nil {
			panic(m.panic)
		}
		return m.val
	}
	go func() {
		defer func() {
			if running {
				running = false
				cout <- msg[Out]{panic: recover()}
			}
		}()
		var out Out
		m := <-cin
		if m.panic == nil {
			out = f(m.val, yield)
		}
		running = false
		cout <- msg[Out]{val: out}
	}()
	return resume, cancel
}

func TaskIterator[V any](push func(yield func(V) bool)) (pull func() (V, bool), stop func() V, cancel func()) {
	copush := func(more bool, yield func(V) bool) V {
		if more {
			push(yield)
		}
		var zero V
		return zero
	}
	resume, cancel := newCoroutine(copush)
	pull = func() (V, bool) {
		return resume(true)
	}
	stop = func() V {
		v, _ := resume(false)
		fmt.Printf("stop called: %v\n", v)
		return v
	}
	return pull, stop, cancel
}
