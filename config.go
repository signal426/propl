package protopolicy

type Config[T any] struct {
	authorizer T
}
