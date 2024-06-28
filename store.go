package protopolicy

import (
	"fmt"
	"reflect"

	"github.com/signal426/protopolicy/policy"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type (
	fieldStore map[string]*fieldData
)

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

var _ policy.Subject = (*fieldData)(nil)

type fieldData struct {
	zero   bool
	val    any
	path   string
	inMask bool
}

// HasTrait implements policy.Subject.
func (f *fieldData) HasTrait(t policy.Trait) bool {
	if t.Trait() == policy.NotZero && f == nil || f.zero {
		return false
	}
	if t.Trait() == policy.NotEq {
		// todo
	}
	if t.Trait() == policy.Calculated {
		t.Calculate(f.val)
	}
	return true
}

// MeetsConditions implements policy.Subject.
func (f *fieldData) MeetsConditions(conditions policy.Condition) bool {
	if f == nil {
		if conditions.Has(policy.InMessage) {
			return false
		}
		if conditions.Has(policy.InMask) {
			return false
		}
	}
	return true
}

func newFieldData(value protoreflect.Value, inMask bool, name, parent, delimeter string) *fieldData {
	if parent != "" {
		name = fmt.Sprintf("%s%s%s", parent, delimeter, name)
	}
	return &fieldData{
		zero:   !value.IsValid() || reflect.ValueOf(value).IsZero(),
		val:    value,
		path:   name,
		inMask: inMask,
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

func messageToFieldStore(message proto.Message, delimeter string, paths ...string) fieldStore {
	var pathSet PathSet
	if len(paths) > 0 {
		pathSet = NewPathSet(paths...)
	}
	return traverseMessageForFieldStore(message, pathSet, nil, true, "", delimeter)
}

func traverseMessageForFieldStore(message proto.Message, paths PathSet, store fieldStore, init bool, parent string, delimeter string) fieldStore {
	fmt.Printf("store: %+v\n", store)
	if message == nil || store.empty() && !init {
		return nil
	}
	if init {
		init = false
		store = make(map[string]*fieldData)
	}
	if message.ProtoReflect().Descriptor().Fields().Len() == 0 {
		return store
	}
	message.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		fieldValue := message.ProtoReflect().Get(fd)
		var inMask bool
		if paths != nil {
			if !paths.Has(string(fd.Name())) {
				inMask = fd.HasJSONName() && paths.Has(fd.JSONName())
			} else {
				inMask = true
			}
		}
		fieldData := newFieldData(fieldValue, inMask, string(fd.Name()), parent, delimeter)
		store.add(fieldData)
		if fd.Message() == nil {
			fmt.Printf("got here\n")
			return true
		}
		fmt.Printf("%+v\n", fieldValue.Message().Interface())
		traverseMessageForFieldStore(fieldValue.Message().Interface(), paths, store, init, fieldData.p(), delimeter)
		return true
	})
	return store
}
