package async

import (
	"context"
)

type Awaiter interface {
	// Add add an action
	Add(action Action)
	// Wait wail for all actions to completed
	Wait(context.Context) ([]error, error)
	// WaitAny wait for any action to completed without error, can cancel other tasks
	WaitAny(context.Context) ([]error, error)
	// WaitN wait for N actions to completed without error
	WaitN(context.Context, int) ([]error, error)
}

type awaiter struct {
	actions []Action
}

func (a *awaiter) Add(action Action) {
	a.actions = append(a.actions, action)
}

func (a *awaiter) Wait(ctx context.Context) ([]error, error) {
	wait := make(chan error)

	for _, action := range a.actions {
		go func(action Action) {

			wait <- action(ctx)
		}(action)
	}

	var taskErrs []error

	tt := len(a.actions)
	for i := 0; i < tt; i++ {
		select {
		case err := <-wait:
			if err != nil {
				taskErrs = append(taskErrs, err)
			}
		case <-ctx.Done():
			return taskErrs, ctx.Err()
		}
	}

	if len(taskErrs) > 0 {
		return taskErrs, ErrTooLessDone
	}

	return taskErrs, nil
}

func (a *awaiter) WaitN(ctx context.Context, n int) ([]error, error) {
	wait := make(chan error)

	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, action := range a.actions {
		go func(action Action) {
			wait <- action(cancelCtx)

		}(action)
	}

	var taskErrs []error
	tt := len(a.actions)

	var done int
	for i := 0; i < tt; i++ {
		select {
		case err := <-wait:
			if err != nil {
				taskErrs = append(taskErrs, err)
			} else {

				done++
				if done == n {
					return taskErrs, nil
				}
			}
		case <-ctx.Done():
			return taskErrs, ctx.Err()
		}

	}

	return taskErrs, ErrTooLessDone
}

func (a *awaiter) WaitAny(ctx context.Context) ([]error, error) {
	return a.WaitN(ctx, 1)
}
