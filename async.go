package async

import (
	"context"
	"errors"
)

var (
	ErrTooLessDone = errors.New("async: too less tasks/actions to completed without error")
)

// Task a task with result T
type Task[T any] func(ctx context.Context) (T, error)

// New create a task waiter
func New[T any](tasks ...Task[T]) Waiter[T] {
	return &waiter[T]{
		tasks: tasks,
	}
}

// Action a task without result
type Action func(ctx context.Context) error

// NewA create an action awaiter
func NewA(actions ...Action) Awaiter {
	return &awaiter{
		actions: actions,
	}
}
