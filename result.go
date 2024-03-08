package async

type Result[T any] struct {
	Data  T
	Error error
}
