package protovalidate

import "google.golang.org/protobuf/proto"

func FieldNeverZero[T proto.Message](id string, value any) *fieldPolicy[T] {
	return newFieldPolicy[T](id, NeverZero, value, nil)
}
