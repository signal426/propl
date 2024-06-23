package protovalidate

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type fieldStore map[string]*fieldData

func (f fieldStore) Empty() bool {
	return f == nil || len(f) == 0
}

type fieldData struct {
	zero bool
	val  any
	path string
}

func (f *fieldData) Zero() bool {
	return f.zero
}

func (f *fieldData) Value() any {
	return f.val
}

func (f *fieldData) Path() string {
	return f.path
}

func MessageToFieldStore(message proto.Message, delimeter string) fieldStore {
	return messageToFieldStore(message, nil, true, "", delimeter)
}

func messageToFieldStore(message proto.Message, store fieldStore, init bool, parent string, delimeter string) fieldStore {
	if message == nil || store.Empty() && !init {
		return nil
	}
	if store == nil {
		init = false
		store = make(fieldStore)
	}
	message.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if fd.Name().IsValid() {
			value := message.ProtoReflect().Get(fd)
			path := string(fd.Name())
			if parent != "" {
				path = fmt.Sprintf("%s%s%s", parent, delimeter, path)
			}
			store[string(fd.Name())] = &fieldData{
				// implied at this point that
				zero: !value.IsValid(),
				val:  value,
				path: path,
			}
			if fd.Message() != nil {
				fieldDescriptorsToStore(value.Message().Interface(), store, init, path, delimeter)
			}
		}
		return true
	})
	return store
}
