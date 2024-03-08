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
	var wantedErrs error = Errors([]error{wantedErr})

	tests := []struct {
		name         string
		ctx          func() context.Context
		withCancel   bool
		setup        func() Awaiter[int]
		wantedResult []int
		wantedError  error
	}{
		{
			name: "wait_should_work",
			ctx:  func() context.Context { return context.Background() },
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
			wantedError:  nil,
		},
		{
			name: "error_should_work",
			ctx:  func() context.Context { return context.Background() },
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
			wantedError:  wantedErrs,
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
			wantedError:  context.DeadlineExceeded,
		},
		{
			name: "cancel_should_work",
			ctx: func() context.Context {
				return context.Background()
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
			withCancel:   true,
			wantedResult: []int{3},
			wantedError:  context.Canceled,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := test.setup()
			var result []int
			var err error

			if test.withCancel {
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(1 * time.Second)
					cancel()
				}()
				result, err = a.Wait(ctx)
			} else {
				result, err = a.Wait(test.ctx())
			}

			slices.Sort(result)

			require.Equal(t, test.wantedResult, result)
			require.Equal(t, test.wantedError, err)

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
		wantedError  error
	}{
		{
			name: "1st_should_work",
			ctx: func() context.Context {
				return context.Background()
			},
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
			ctx: func() context.Context {
				return context.Background()
			},
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
			ctx: func() context.Context {
				return context.Background()
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
			wantedResult: 3,
		},
		{
			name: "slowest_should_work",
			ctx:  func() context.Context { return context.Background() },
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
		},
		{
			name: "fastest_should_work",
			ctx:  func() context.Context { return context.Background() },
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
			ctx:  func() context.Context { return context.Background() },
			setup: func() Awaiter[int] {
				return New[int](func(ctx context.Context) (int, error) {
					return 0, wantedErr
				}, func(ctx context.Context) (int, error) {
					return 0, wantedErr
				}, func(ctx context.Context) (int, error) {
					return 0, wantedErr
				})
			},
			wantedError: Errors([]error{wantedErr, wantedErr, wantedErr}),
		},
		{
			name: "error_should_work",
			ctx:  func() context.Context { return context.Background() },
			setup: func() Awaiter[int] {
				return New[int](func(ctx context.Context) (int, error) {
					return 0, wantedErr
				})
			},
			wantedError: Errors([]error{wantedErr}),
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
			wantedError: context.DeadlineExceeded,
		},
		{
			name: "cancel_should_work",
			ctx: func() context.Context {
				return context.Background()
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
			withCancel:  true,
			wantedError: context.Canceled,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := test.setup()
			var result int
			var err error

			if test.withCancel {
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(1 * time.Second)
					cancel()
				}()
				result, err = a.WaitAny(ctx)
			} else {
				result, err = a.WaitAny(test.ctx())
			}

			require.Equal(t, test.wantedResult, result)
			require.Equal(t, test.wantedError, err)
		})

	}
}
