package async

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAwait(t *testing.T) {

	wantedErr := errors.New("wanted")

	tests := []struct {
		name       string
		ctx        func() context.Context
		withCancel bool
		setup      func() Awaiter
		wantedErr  error
		wantedErrs []error
	}{
		{
			name: "wait_should_work",
			ctx:  context.Background,
			setup: func() Awaiter {
				a := NewA(func(ctx context.Context) error {
					return nil
				}, func(ctx context.Context) error {
					return nil
				})

				a.Add(func(ctx context.Context) error {
					return nil
				})

				return a
			},

			wantedErr: nil,
		},
		{
			name: "error_should_work",
			ctx:  context.Background,
			setup: func() Awaiter {
				return NewA(func(ctx context.Context) error {
					return nil
				}, func(ctx context.Context) error {
					return nil
				}, func(ctx context.Context) error {
					return wantedErr
				})
			},
			wantedErr:  ErrTooLessDone,
			wantedErrs: []error{wantedErr},
		},
		{
			name: "errors_should_work",
			ctx:  context.Background,
			setup: func() Awaiter {
				return NewA(func(ctx context.Context) error {
					return wantedErr
				}, func(ctx context.Context) error {
					return wantedErr
				}, func(ctx context.Context) error {
					return wantedErr
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
			setup: func() Awaiter {
				return NewA(func(ctx context.Context) error {
					time.Sleep(5 * time.Second)
					return nil
				}, func(ctx context.Context) error {
					time.Sleep(5 * time.Second)
					return nil
				}, func(ctx context.Context) error {
					return nil
				})
			},

			wantedErr: context.DeadlineExceeded,
		},
		{
			name: "cancel_should_work",
			ctx:  context.Background,
			setup: func() Awaiter {
				return NewA(func(ctx context.Context) error {
					time.Sleep(5 * time.Second)
					return nil
				}, func(ctx context.Context) error {
					time.Sleep(5 * time.Second)
					return nil
				}, func(ctx context.Context) error {
					return nil
				})
			},
			withCancel: true,
			wantedErr:  context.Canceled,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := test.setup()

			var err error
			var taskErrs []error

			if test.withCancel {
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(1 * time.Second)
					cancel()
				}()
				taskErrs, err = a.Wait(ctx)
			} else {
				taskErrs, err = a.Wait(test.ctx())
			}

			require.Equal(t, test.wantedErr, err)
			require.Equal(t, test.wantedErrs, taskErrs)

		})

	}
}

func TestAwaitAny(t *testing.T) {

	wantedErr := errors.New("wanted")

	tests := []struct {
		name       string
		ctx        func() context.Context
		withCancel bool
		setup      func() Awaiter
		wantedErr  error
		wantedErrs []error
	}{
		{
			name: "1st_should_work",
			ctx:  context.Background,
			setup: func() Awaiter {
				a := NewA(func(ctx context.Context) error {
					time.Sleep(5 * time.Second)
					return nil
				}, func(ctx context.Context) error {
					time.Sleep(5 * time.Second)
					return nil
				})

				a.Add(func(ctx context.Context) error {
					return nil
				})

				return a
			},
		},
		{
			name: "2nd_should_work",
			ctx:  context.Background,
			setup: func() Awaiter {
				return NewA(func(ctx context.Context) error {
					time.Sleep(5 * time.Second)
					return nil
				}, func(ctx context.Context) error {
					return nil
				}, func(ctx context.Context) error {
					time.Sleep(5 * time.Second)
					return nil
				})
			},
		},
		{
			name: "3rd_should_work",
			ctx:  context.Background,
			setup: func() Awaiter {
				return NewA(func(ctx context.Context) error {
					time.Sleep(5 * time.Second)
					return nil
				}, func(ctx context.Context) error {
					time.Sleep(5 * time.Second)
					return nil
				}, func(ctx context.Context) error {

					return nil
				})
			},
		},
		{
			name: "slowest_should_work",
			ctx:  context.Background,
			setup: func() Awaiter {
				return NewA(func(ctx context.Context) error {
					time.Sleep(3 * time.Second)
					return nil
				}, func(ctx context.Context) error {
					return wantedErr
				}, func(ctx context.Context) error {
					return wantedErr
				})
			},
			wantedErrs: []error{wantedErr, wantedErr},
		},
		{
			name: "fastest_should_work",
			ctx:  context.Background,
			setup: func() Awaiter {
				return NewA(func(ctx context.Context) error {
					return nil
				}, func(ctx context.Context) error {
					time.Sleep(2 * time.Second)
					return wantedErr
				}, func(ctx context.Context) error {
					time.Sleep(3 * time.Second)
					return wantedErr
				})
			},
		},
		{
			name: "errors_should_work",
			ctx:  context.Background,
			setup: func() Awaiter {
				return NewA(func(ctx context.Context) error {
					return wantedErr
				}, func(ctx context.Context) error {
					return wantedErr
				}, func(ctx context.Context) error {
					return wantedErr
				})
			},
			wantedErr:  ErrTooLessDone,
			wantedErrs: []error{wantedErr, wantedErr, wantedErr},
		},
		{
			name: "error_should_work",
			ctx:  context.Background,
			setup: func() Awaiter {
				return NewA(func(ctx context.Context) error {
					return wantedErr
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
			setup: func() Awaiter {
				return NewA(func(ctx context.Context) error {
					time.Sleep(5 * time.Second)
					return nil
				}, func(ctx context.Context) error {
					time.Sleep(5 * time.Second)
					return nil
				}, func(ctx context.Context) error {
					time.Sleep(5 * time.Second)
					return nil
				})
			},
			wantedErr: context.DeadlineExceeded,
		},
		{
			name: "cancel_should_work",
			ctx:  context.Background,
			setup: func() Awaiter {
				return NewA(func(ctx context.Context) error {
					time.Sleep(5 * time.Second)
					return nil
				}, func(ctx context.Context) error {
					time.Sleep(5 * time.Second)
					return nil
				}, func(ctx context.Context) error {
					time.Sleep(5 * time.Second)
					return nil
				})
			},
			withCancel: true,
			wantedErr:  context.Canceled,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := test.setup()

			var err error
			var taskErrs []error

			if test.withCancel {
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(1 * time.Second)
					cancel()
				}()
				taskErrs, err = a.WaitAny(ctx)
			} else {
				taskErrs, err = a.WaitAny(test.ctx())
			}

			require.Equal(t, test.wantedErr, err)
			require.Equal(t, test.wantedErrs, taskErrs)
		})

	}
}

func TestAwaitN(t *testing.T) {

	wantedErr := errors.New("wanted")

	tests := []struct {
		name       string
		ctx        func() context.Context
		withCancel bool
		setup      func() Awaiter
		wantedN    int
		wantedErr  error
		wantedErrs []error
	}{
		{
			name: "wait_n_should_work",
			ctx:  context.Background,
			setup: func() Awaiter {
				a := NewA(func(ctx context.Context) error {
					return nil
				}, func(ctx context.Context) error {
					return nil
				})

				a.Add(func(ctx context.Context) error {
					time.Sleep(1 * time.Second)
					return nil
				})

				return a
			},
			wantedN: 2,
		},
		{
			name: "error_n_should_work",
			ctx:  context.Background,
			setup: func() Awaiter {
				return NewA(func(ctx context.Context) error {
					time.Sleep(1 * time.Second)
					return nil
				}, func(ctx context.Context) error {
					time.Sleep(1 * time.Second)
					return nil
				}, func(ctx context.Context) error {
					return wantedErr
				})
			},
			wantedN:    2,
			wantedErrs: []error{wantedErr},
		},
		{
			name: "context_should_work",
			ctx: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 3*time.Second) //nolint
				return ctx
			},
			setup: func() Awaiter {
				return NewA(func(ctx context.Context) error {
					time.Sleep(5 * time.Second)
					return nil
				}, func(ctx context.Context) error {
					time.Sleep(5 * time.Second)
					return nil
				}, func(ctx context.Context) error {
					return nil
				})
			},
			wantedErr: context.DeadlineExceeded,
		},
		{
			name: "cancel_should_work",
			ctx:  context.Background,
			setup: func() Awaiter {
				return NewA(func(ctx context.Context) error {
					time.Sleep(5 * time.Second)
					return nil
				}, func(ctx context.Context) error {
					time.Sleep(5 * time.Second)
					return nil
				}, func(ctx context.Context) error {
					return nil
				})
			},
			withCancel: true,

			wantedErr: context.Canceled,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := test.setup()
			var err error
			var taskErrs []error

			if test.withCancel {
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(1 * time.Second)
					cancel()
				}()
				taskErrs, err = a.WaitN(ctx, test.wantedN)
			} else {
				taskErrs, err = a.WaitN(test.ctx(), test.wantedN)
			}

			require.Equal(t, test.wantedErr, err)
			require.Equal(t, test.wantedErrs, taskErrs)

		})

	}
}
