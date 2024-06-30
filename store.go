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
		// we may have a field that was in a mask, but not set in the message.
		// in this case, we don't know the path, so try fetching by its
		// name.
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
func (f *fieldData) HasTrait(t trait) bool {
	if t.traitType == notZero && f.z() {
		return false
	}
	if t.traitType == calculated {
		return t.calculate(f.v())
	}
	return true
}

// ConditionalAction implements policy.Subject.
func (f *fieldData) ConditionalAction(conditions Condition) Action {
	if !f.s() && conditions.Has(InMessage) {
		return Fail
	}
	if !f.s() && conditions.Has(InMask) && f.m() {
		return Fail
	}
	if !conditions.Has(InMessage) && conditions.Has(InMask) && !f.m() {
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

// fill fills the store with field data from the request message. It keeps track
// of the mask paths requested (if present), and creates unset fields for anything
// specified in the mask but not set on the message.
func (store fieldStore) fill(message proto.Message, paths ...string) {
	pathSet := newPathSet(paths...)
	fillStore(message, pathSet, store, true, "")
	// if we have unclaimed paths, create unset fields from them
	for _, uc := range pathSet.unclaimed() {
		store.add(newUnsetFieldData(uc, true))
	}
}

// fillStore recursively ranges over the fields in the message.
func fillStore(message proto.Message, paths pathSet, store fieldStore, init bool, parent string) {
	if message == nil || store.empty() && !init {
		return
	}
	message.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		fieldValue := message.ProtoReflect().Get(fd)
		var (
			inMask bool
			match  string
		)
		if !paths.empty() {
			if paths.has(string(fd.Name())) {
				inMask = true
				match = string(fd.Name())
			} else if fd.HasJSONName() && paths.has(fd.JSONName()) {
				inMask = true
				match = fd.JSONName()
			}
			// if it's in the mask, claim it
			if inMask {
				paths.claim(match)
			}
		}
		fieldData := newFieldData(fieldValue.Interface(), fieldValue.IsValid(), inMask, string(fd.Name()), parent)
		store.add(fieldData)
		if fd.Message() == nil {
			return true
		}
		fillStore(fieldValue.Message().Interface(), paths, store, false, fieldData.p())
		return true
	})
	return
}
