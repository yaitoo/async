package async

import (
	"context"
)

type Awaiter[T any] interface {
	Add(task func(context.Context) (T, error))
	Wait(context.Context) ([]T, error)
	WaitAny(context.Context) (T, error)
}

type awaiter[T any] struct {
	tasks []func(context.Context) (T, error)
}

func (a *awaiter[T]) Add(task func(ctx context.Context) (T, error)) {
	a.tasks = append(a.tasks, task)
}

func (a *awaiter[T]) Wait(ctx context.Context) ([]T, error) {
	wait := make(chan Result[T])

	n := len(a.tasks)
	for _, task := range a.tasks {
		go func(task func(context.Context) (T, error)) {
			r, err := task(ctx)
			wait <- Result[T]{
				Data:  r,
				Error: err,
			}
		}(task)
	}

	var r Result[T]
	var es Errors
	var items []T

	for i := 0; i < n; i++ {
		select {
		case r = <-wait:
			if r.Error != nil {
				es = append(es, r.Error)
			} else {
				items = append(items, r.Data)
			}
		case <-ctx.Done():
			return items, ctx.Err()
		}

	}

	if len(es) > 0 {
		return items, es
	}

	return items, nil
}

func (a *awaiter[T]) WaitAny(ctx context.Context) (T, error) {

	n := len(a.tasks)

	wait := make(chan Result[T])
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, task := range a.tasks {
		go func(task func(context.Context) (T, error)) {
			r, err := task(cancelCtx)
			wait <- Result[T]{
				Data:  r,
				Error: err,
			}
		}(task)
	}

	var t T

	var r Result[T]
	var es Errors

	for i := 0; i < n; i++ {
		select {
		case r = <-wait:
			if r.Error == nil {
				return r.Data, nil
			}

			es = append(es, r.Error)
		case <-ctx.Done():
			return t, ctx.Err()
		}
	}

	if len(es) > 0 {
		return t, es
	}

	return t, nil
}
