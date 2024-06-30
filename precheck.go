package propl

import "google.golang.org/protobuf/proto"

type Precheck[T proto.Message] func(msg T) error
