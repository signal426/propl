package protopolicy

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type fieldStore map[string]*fieldData

func (f fieldStore) empty() bool {
	return f == nil || len(f) == 0
}

func (f fieldStore) add(fd *fieldData) {
	f[fd.p()] = fd
}

func (f fieldStore) getByPath(p string) *fieldData {
	fd, ok := f[p]
	if !ok {
		return nil
	}
	return fd
}

type fieldData struct {
	zero bool
	val  any
	path string
}

func newFieldData(field protoreflect.FieldDescriptor, value protoreflect.Value, parent, delimeter string) *fieldData {
	if !field.Name().IsValid() {
		return nil
	}
	path := string(field.Name())
	if parent != "" {
		path = fmt.Sprintf("%s%s%s", parent, delimeter, path)
	}
	return &fieldData{
		zero: !value.IsValid(),
		val:  value,
		path: path,
	}
}

func (f fieldData) z() bool {
	return f.zero
}

func (f fieldData) v() any {
	return f.val
}

func (f fieldData) p() string {
	return f.path
}

func messageToFieldStore(message proto.Message, delimeter string) fieldStore {
	return traverseMessageForFieldStore(message, nil, true, "", delimeter)
}

func traverseMessageForFieldStore(message proto.Message, store fieldStore, init bool, parent string, delimeter string) fieldStore {
	if message == nil || store.empty() && !init {
		return nil
	}
	if init {
		init = false
		store = make(fieldStore)
	}
	if message.ProtoReflect().Descriptor().Fields().Len() == 0 {
		return store
	}
	message.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		fieldValue := message.ProtoReflect().Get(fd)
		fieldData := newFieldData(fd, fieldValue, parent, delimeter)
		if fieldData == nil {
			// keep ranging
			return true
		}
		store.add(fieldData)
		traverseMessageForFieldStore(fieldValue.Message().Interface(), store, init, fieldData.p(), delimeter)
		return true
	})
	return store
}
