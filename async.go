package async

import "context"

func New[T any](tasks ...func(ctx context.Context) (T, error)) Awaiter[T] {
	return &awaiter[T]{
		tasks: tasks,
	}
}
