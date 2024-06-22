package protovalidate

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type fieldStore map[string]*fieldData

type fieldData struct {
	zero bool
	val  any
	loc  string
}

func (f *fieldData) Zero() bool {
	return f.zero
}

func (f *fieldData) Value() any {
	return f.val
}

func (f *fieldData) Loc() string {
	return f.loc
}

func FieldDescriptorsAsStore(message proto.Message) fieldStore {
	return fieldDescriptorsAsPathMap(message, nil, true, "")
}

func fieldDescriptorsAsPathMap(message proto.Message, fields []protoreflect.FieldDescriptor, init bool, parent string) fieldStore {
	if len(fields) == 0 && !init {
		return nil
	}
	store := make(fieldStore)
	message.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if fd.Name().IsValid() {
			store[string(fd.Name())] = &fieldData{
				// implied at this point that
				zero: fd.HasPresence(),
			}
		}
		return true
	})
}
