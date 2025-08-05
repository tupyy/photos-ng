package processing

import (
	"context"
	"fmt"
	"iter"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
)

type Worker[T, R any] struct {
	fn          func(ctx context.Context, t T) (R, error)
	input       iter.Seq[T]
	stopOnError bool
}

func NewWorker[T, R any](input iter.Seq[T], fn func(ctx context.Context, t T) (R, error)) *Worker[T, R] {
	return &Worker[T, R]{
		fn: fn,
	}
}

func (s *Worker[T, R]) StopOnError(stopOnError bool) {
	s.stopOnError = stopOnError
}

func (s *Worker[T, R]) Run(ctx context.Context) error {
	for input := range s.input {
		if input.Err != nil {
			s.output <- entity.Result[R]{Err: input.Err}
			continue
		}

		r, err := s.fn(ctx, input.Data)

		if err != nil {
			if s.stopOnError {
				return fmt.Errorf("Worker %s failed: %w", s.name, err)
			}
			s.output <- entity.Result[R]{Err: err}
			continue
		}

		s.output <- entity.Result[R]{Data: r}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	return nil
}
