package protovalidate

import (
	"google.golang.org/protobuf/proto"
)

type NoAuth struct{}

type authzFn[T proto.Message, U any] func(t T, u any) error

type authzPolicy[T proto.Message, U any] struct {
	fieldMeta *fieldMeta
	authzFn   authzFn[T, U]
}

func newAuthzPolicy[T proto.Message, U any](id string, authzFn authzFn[T, U]) *authzPolicy[T, U] {
	return &authzPolicy[T, U]{
		fieldMeta: newFieldMeta(id),
		authzFn:   authzFn,
	}
}
