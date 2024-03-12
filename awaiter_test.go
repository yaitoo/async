package async

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWait(t *testing.T) {

	wantedErr := errors.New("wanted")
	wantedErrs := []error{wantedErr}

	tests := []struct {
		name         string
		ctx          func() context.Context
		withCancel   bool
		setup        func() Awaiter[int]
		wantedResult []int
		wantedErr    error
		wantedErrs   []error
	}{
		{
			name: "wait_should_work",
			ctx:  context.Background,
			setup: func() Awaiter[int] {
				a := New[int](func(ctx context.Context) (int, error) {
					return 1, nil
				}, func(ctx context.Context) (int, error) {
					return 2, nil
				})

				a.Add(func(ctx context.Context) (int, error) {
					return 3, nil
				})

				return a
			},
			wantedResult: []int{1, 2, 3},
			wantedErr:    nil,
		},
		{
			name: "error_should_work",
			ctx:  context.Background,
			setup: func() Awaiter[int] {
				return New[int](func(ctx context.Context) (int, error) {
					return 1, nil
				}, func(ctx context.Context) (int, error) {
					return 2, nil
				}, func(ctx context.Context) (int, error) {
					return 0, wantedErr
				})
			},
			wantedResult: []int{1, 2},
			wantedErr:    ErrTooLessDone,
			wantedErrs:   wantedErrs,
		},
		{
			name: "errors_should_work",
			ctx:  context.Background,
			setup: func() Awaiter[int] {
				return New[int](func(ctx context.Context) (int, error) {
					return 0, wantedErr
				}, func(ctx context.Context) (int, error) {
					return 0, wantedErr
				}, func(ctx context.Context) (int, error) {
					return 0, wantedErr
				})
			},
			wantedErr:  ErrTooLessDone,
			wantedErrs: []error{wantedErr, wantedErr, wantedErr},
		},
		{
			name: "context_should_work",
			ctx: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 3*time.Second) //nolint
				return ctx
			},
			setup: func() Awaiter[int] {
				return New[int](func(ctx context.Context) (int, error) {
					time.Sleep(5 * time.Second)
					return 1, nil
				}, func(ctx context.Context) (int, error) {
					time.Sleep(5 * time.Second)
					return 2, nil
				}, func(ctx context.Context) (int, error) {
					return 3, nil
				})
			},
			wantedResult: []int{3},
			wantedErr:    context.DeadlineExceeded,
		},
		{
			name: "cancel_should_work",
			ctx:  context.Background,
			setup: func() Awaiter[int] {
				return New[int](func(ctx context.Context) (int, error) {
					time.Sleep(5 * time.Second)
					return 1, nil
				}, func(ctx context.Context) (int, error) {
					time.Sleep(5 * time.Second)
					return 2, nil
				}, func(ctx context.Context) (int, error) {
					return 3, nil
				})
			},
			withCancel:   true,
			wantedResult: []int{3},
			wantedErr:    context.Canceled,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := test.setup()
			var result []int
			var err error
			var taskErrs []error

			if test.withCancel {
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(1 * time.Second)
					cancel()
				}()
				result, taskErrs, err = a.Wait(ctx)
			} else {
				result, taskErrs, err = a.Wait(test.ctx())
			}

			slices.Sort(result)

			require.Equal(t, test.wantedResult, result)
			require.Equal(t, test.wantedErr, err)
			require.Equal(t, test.wantedErrs, taskErrs)

		})

	}
}

func TestWaitAny(t *testing.T) {

	wantedErr := errors.New("wanted")

	tests := []struct {
		name         string
		ctx          func() context.Context
		withCancel   bool
		setup        func() Awaiter[int]
		wantedResult int
		wantedErr    error
		wantedErrs   []error
	}{
		{
			name: "1st_should_work",
			ctx:  context.Background,
			setup: func() Awaiter[int] {
				a := New[int](func(ctx context.Context) (int, error) {
					time.Sleep(5 * time.Second)
					return 2, nil
				}, func(ctx context.Context) (int, error) {
					time.Sleep(5 * time.Second)
					return 3, nil
				})

				a.Add(func(ctx context.Context) (int, error) {
					return 1, nil
				})

				return a
			},
			wantedResult: 1,
		},
		{
			name: "2nd_should_work",
			ctx:  context.Background,
			setup: func() Awaiter[int] {
				return New[int](func(ctx context.Context) (int, error) {
					time.Sleep(5 * time.Second)
					return 1, nil
				}, func(ctx context.Context) (int, error) {
					return 2, nil
				}, func(ctx context.Context) (int, error) {
					time.Sleep(5 * time.Second)
					return 3, nil
				})
			},
			wantedResult: 2,
		},
		{
			name: "3rd_should_work",
			ctx:  context.Background,
			setup: func() Awaiter[int] {
				return New[int](func(ctx context.Context) (int, error) {
					time.Sleep(5 * time.Second)
					return 1, nil
				}, func(ctx context.Context) (int, error) {
					time.Sleep(5 * time.Second)
					return 2, nil
				}, func(ctx context.Context) (int, error) {

					return 3, nil
				})
			},
			wantedResult: 3,
		},
		{
			name: "slowest_should_work",
			ctx:  context.Background,
			setup: func() Awaiter[int] {
				return New[int](func(ctx context.Context) (int, error) {
					time.Sleep(3 * time.Second)
					return 1, nil
				}, func(ctx context.Context) (int, error) {
					return 0, wantedErr
				}, func(ctx context.Context) (int, error) {
					return 0, wantedErr
				})
			},
			wantedResult: 1,
			wantedErrs:   []error{wantedErr, wantedErr},
		},
		{
			name: "fastest_should_work",
			ctx:  context.Background,
			setup: func() Awaiter[int] {
				return New[int](func(ctx context.Context) (int, error) {
					return 1, nil
				}, func(ctx context.Context) (int, error) {
					time.Sleep(2 * time.Second)
					return 0, wantedErr
				}, func(ctx context.Context) (int, error) {
					time.Sleep(3 * time.Second)
					return 0, wantedErr
				})
			},
			wantedResult: 1,
		},
		{
			name: "errors_should_work",
			ctx:  context.Background,
			setup: func() Awaiter[int] {
				return New[int](func(ctx context.Context) (int, error) {
					return 0, wantedErr
				}, func(ctx context.Context) (int, error) {
					return 0, wantedErr
				}, func(ctx context.Context) (int, error) {
					return 0, wantedErr
				})
			},
			wantedErr:  ErrTooLessDone,
			wantedErrs: []error{wantedErr, wantedErr, wantedErr},
		},
		{
			name: "error_should_work",
			ctx:  context.Background,
			setup: func() Awaiter[int] {
				return New[int](func(ctx context.Context) (int, error) {
					return 0, wantedErr
				})
			},
			wantedErr:  ErrTooLessDone,
			wantedErrs: []error{wantedErr},
		},
		{
			name: "context_should_work",
			ctx: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 3*time.Second) //nolint
				return ctx
			},
			setup: func() Awaiter[int] {
				return New[int](func(ctx context.Context) (int, error) {
					time.Sleep(5 * time.Second)
					return 1, nil
				}, func(ctx context.Context) (int, error) {
					time.Sleep(5 * time.Second)
					return 2, nil
				}, func(ctx context.Context) (int, error) {
					time.Sleep(5 * time.Second)
					return 3, nil
				})
			},
			wantedErr: context.DeadlineExceeded,
		},
		{
			name: "cancel_should_work",
			ctx:  context.Background,
			setup: func() Awaiter[int] {
				return New[int](func(ctx context.Context) (int, error) {
					time.Sleep(5 * time.Second)
					return 1, nil
				}, func(ctx context.Context) (int, error) {
					time.Sleep(5 * time.Second)
					return 2, nil
				}, func(ctx context.Context) (int, error) {
					time.Sleep(5 * time.Second)
					return 3, nil
				})
			},
			withCancel: true,
			wantedErr:  context.Canceled,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := test.setup()
			var result int
			var err error
			var taskErrs []error

			if test.withCancel {
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(1 * time.Second)
					cancel()
				}()
				result, taskErrs, err = a.WaitAny(ctx)
			} else {
				result, taskErrs, err = a.WaitAny(test.ctx())
			}

			require.Equal(t, test.wantedResult, result)
			require.Equal(t, test.wantedErr, err)
			require.Equal(t, test.wantedErrs, taskErrs)
		})

	}
}

func TestWaitN(t *testing.T) {

	wantedErr := errors.New("wanted")

	tests := []struct {
		name         string
		ctx          func() context.Context
		withCancel   bool
		setup        func() Awaiter[int]
		wantedN      int
		wantedResult []int
		wantedErr    error
		wantedErrs   []error
	}{
		{
			name: "wait_n_should_work",
			ctx:  context.Background,
			setup: func() Awaiter[int] {
				a := New[int](func(ctx context.Context) (int, error) {
					return 1, nil
				}, func(ctx context.Context) (int, error) {
					return 2, nil
				})

				a.Add(func(ctx context.Context) (int, error) {
					time.Sleep(1 * time.Second)
					return 3, nil
				})

				return a
			},
			wantedN:      2,
			wantedResult: []int{1, 2},
		},
		{
			name: "error_n_should_work",
			ctx:  context.Background,
			setup: func() Awaiter[int] {
				return New[int](func(ctx context.Context) (int, error) {
					time.Sleep(1 * time.Second)
					return 1, nil
				}, func(ctx context.Context) (int, error) {
					time.Sleep(1 * time.Second)
					return 2, nil
				}, func(ctx context.Context) (int, error) {
					return 0, wantedErr
				})
			},
			wantedN:      2,
			wantedResult: []int{1, 2},
			wantedErrs:   []error{wantedErr},
		},
		{
			name: "context_should_work",
			ctx: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 3*time.Second) //nolint
				return ctx
			},
			setup: func() Awaiter[int] {
				return New[int](func(ctx context.Context) (int, error) {
					time.Sleep(5 * time.Second)
					return 1, nil
				}, func(ctx context.Context) (int, error) {
					time.Sleep(5 * time.Second)
					return 2, nil
				}, func(ctx context.Context) (int, error) {
					return 3, nil
				})
			},
			wantedResult: []int{3},
			wantedErr:    context.DeadlineExceeded,
		},
		{
			name: "cancel_should_work",
			ctx:  context.Background,
			setup: func() Awaiter[int] {
				return New[int](func(ctx context.Context) (int, error) {
					time.Sleep(5 * time.Second)
					return 1, nil
				}, func(ctx context.Context) (int, error) {
					time.Sleep(5 * time.Second)
					return 2, nil
				}, func(ctx context.Context) (int, error) {
					return 3, nil
				})
			},
			withCancel:   true,
			wantedResult: []int{3},
			wantedErr:    context.Canceled,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := test.setup()
			var result []int
			var err error
			var taskErrs []error

			if test.withCancel {
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(1 * time.Second)
					cancel()
				}()
				result, taskErrs, err = a.WaitN(ctx, test.wantedN)
			} else {
				result, taskErrs, err = a.WaitN(test.ctx(), test.wantedN)
			}

			slices.Sort(result)

			require.Equal(t, test.wantedResult, result)
			require.Equal(t, test.wantedErr, err)
			require.Equal(t, test.wantedErrs, taskErrs)

		})

	}
}
