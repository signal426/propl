package propl

import (
	"google.golang.org/protobuf/proto"
)

type authorizer[T proto.Message] func(T) error
