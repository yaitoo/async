package async

import (
	"context"
	"errors"
)

var (
	ErrTooLessDone = errors.New("async: too less tasks to completed without error")
)

func New[T any](tasks ...func(ctx context.Context) (T, error)) Awaiter[T] {
	return &awaiter[T]{
		tasks: tasks,
	}
}
