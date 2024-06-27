package protopolicy

import (
	"context"

	"google.golang.org/protobuf/proto"
)

type noAuth struct{}

type authzFn[T proto.Message, U any] func(t T, u any) error

type authzPolicy[T proto.Message, U any] struct {
	authzFn authzFn[T, U]
}

func newAuthzPolicy[T proto.Message, U any](ctx context.Context, authzFn authzFn[T, U]) *authzPolicy[T, U] {
	return &authzPolicy[T, U]{
		authzFn: authzFn,
	}
}
