package async

import (
	"context"
)

type Awaiter[T any] interface {
	// Add add a task
	Add(task func(context.Context) (T, error))
	// Wait wail for all tasks to completed
	Wait(context.Context) ([]T, error, []error)
	// WaitAny wait for any task to completed without error, can cancel other tasks
	WaitAny(context.Context) (T, error, []error)
	// WaitN wait for N tasks to completed without error
	WaitN(context.Context, int) ([]T, error, []error)
}

type awaiter[T any] struct {
	tasks []func(context.Context) (T, error)
}

func (a *awaiter[T]) Add(task func(ctx context.Context) (T, error)) {
	a.tasks = append(a.tasks, task)
}

func (a *awaiter[T]) Wait(ctx context.Context) ([]T, error, []error) {
	wait := make(chan Result[T])

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
	var taskErrs []error
	var items []T

	tt := len(a.tasks)
	for i := 0; i < tt; i++ {
		select {
		case r = <-wait:
			if r.Error != nil {
				taskErrs = append(taskErrs, r.Error)
			} else {
				items = append(items, r.Data)
			}
		case <-ctx.Done():
			return items, ctx.Err(), taskErrs
		}
	}

	return items, nil, taskErrs
}

func (a *awaiter[T]) WaitN(ctx context.Context, n int) ([]T, error, []error) {
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

	var r Result[T]
	var taskErrs []error
	var items []T
	tt := len(a.tasks)
	var done int
	for i := 0; i < tt; i++ {
		select {
		case r = <-wait:
			if r.Error != nil {
				taskErrs = append(taskErrs, r.Error)
			} else {
				items = append(items, r.Data)
				done++
				if done == n {
					return items, nil, taskErrs
				}
			}
		case <-ctx.Done():
			return items, ctx.Err(), taskErrs
		}

	}

	return items, ErrTooLessDone, taskErrs
}

func (a *awaiter[T]) WaitAny(ctx context.Context) (T, error, []error) {
	var t T
	result, err, taskErrs := a.WaitN(ctx, 1)

	if len(result) == 1 {
		t = result[0]
	}

	return t, err, taskErrs
}