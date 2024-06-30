package propl

import (
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type fieldStore map[string]*fieldData

func newFieldStore() fieldStore {
	return make(fieldStore)
}

func (f fieldStore) empty() bool {
	return f == nil || len(f) == 0
}

func (f fieldStore) add(fd *fieldData) {
	f[fd.p()] = fd
}

func (f fieldStore) getByPath(p string) *fieldData {
	fd, ok := f[p]
	if !ok {
		_, name := parseFieldNameFromPath(p)
		fd, _ = f[name]
		return fd
	}
	return fd
}

var _ Subject = (*fieldData)(nil)

type fieldData struct {
	zero   bool
	val    any
	path   string
	inMask bool
	set    bool
}

// HasTrait implements policy.Subject.
func (f *fieldData) HasTrait(t Trait) bool {
	if t.Trait() == notZero && f.z() {
		return false
	}
	if t.Trait() == calculated {
		return t.Calculate(f.v())
	}
	return true
}

// MeetsConditions implements policy.Subject.
func (f *fieldData) ActionFromConditions(conditions Condition) Action {
	if !f.s() && conditions.Has(InMessage) {
		return Fail
	}
	if !f.s() && conditions.Has(InMask) && f.m() {
		return Fail
	}
	if !conditions.Has(InMessage) && conditions.Has(InMask) && !f.m() {
		fmt.Printf("skipping %s\n", f.p())
		return Skip
	}
	return Check
}

func newFieldData(value any, valid, inMask bool, name, parent string) *fieldData {
	if parent != "" {
		name = fmt.Sprintf("%s.%s", parent, name)
	}
	return &fieldData{
		zero:   !valid || reflect.ValueOf(value).IsZero(),
		val:    value,
		path:   name,
		inMask: inMask,
		set:    true,
	}
}

func newUnsetFieldData(name string, inMask bool) *fieldData {
	return &fieldData{
		zero:   true,
		val:    nil,
		path:   name,
		inMask: inMask,
	}
}

func (f fieldData) z() bool {
	return f.zero
}

func (f fieldData) m() bool {
	return f.inMask
}

func (f fieldData) v() any {
	return f.val
}

func (f fieldData) p() string {
	return f.path
}

func (f fieldData) s() bool {
	return f.set
}

func (store fieldStore) fill(message proto.Message, paths ...string) {
	var pathSet PathSet
	if len(paths) > 0 {
		pathSet = NewPathSet(paths...)
	}
	fillStore(message, pathSet, store, true, "")
	if len(pathSet) > 0 {
		for p := range pathSet {
			store.add(newUnsetFieldData(p, true))
		}
	}
}

func fillStore(message proto.Message, paths PathSet, store fieldStore, init bool, parent string) {
	if message == nil || store.empty() && !init {
		return
	}
	message.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		fieldValue := message.ProtoReflect().Get(fd)
		var (
			inMask bool
			match  string
		)
		if paths != nil {
			if paths.Has(string(fd.Name())) {
				inMask = true
				match = string(fd.Name())
			} else if fd.HasJSONName() && paths.Has(fd.JSONName()) {
				inMask = true
				match = fd.JSONName()
			}
			if inMask {
				paths.Remove(match)
			}
		}
		fieldData := newFieldData(fieldValue.Interface(), fieldValue.IsValid(), inMask, string(fd.Name()), parent)
		store.add(fieldData)
		if fieldData == nil || fd.Message() == nil {
			return true
		}
		fillStore(fieldValue.Message().Interface(), paths, store, false, fieldData.p())
		return true
	})
	return
}
