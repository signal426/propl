package propl

import (
	"fmt"
	"reflect"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type fieldStore[T proto.Message] struct {
	msg   T
	store map[string]*fieldData
}

func newFieldStore[T proto.Message](msg T) *fieldStore[T] {
	return &fieldStore[T]{
		msg:   msg,
		store: make(map[string]*fieldData),
	}
}

func (f fieldStore[T]) message() T {
	return f.msg
}

func (f fieldStore[T]) empty() bool {
	return f.store == nil || len(f.store) == 0
}

func (f *fieldStore[T]) add(fd *fieldData) {
	f.store[fd.p()] = fd
}

func (f fieldStore[T]) getByPath(p string) *fieldData {
	fd, ok := f.store[p]
	if !ok {
		// we may have a field that was in a mask, but not set in the message.
		// in this case, we don't know the path, so try fetching by its
		// name.
		_, name := parseFieldNameFromPath(p)
		fd, _ = f.store[name]
		return fd
	}
	return fd
}

var _ Subject = (*fieldData)(nil)

type fieldData struct {
	path   string
	inMask bool
	set    bool
	value  protoreflect.Value
}

// HasTrait implements policy.Subject.
func (f *fieldData) HasTrait(t Trait) bool {
	if t.Type() == NotZero && f.z() {
		return false
	}
	if t.Type() == NotEqual {
		return f.v() != t.NotEqVal()
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

func newFieldData(fv protoreflect.Value, inMask bool, path string) *fieldData {
	return &fieldData{
		value:  fv,
		path:   path,
		inMask: inMask,
		set:    true,
	}
}

func newUnsetFieldData(inMask bool, path string) *fieldData {
	return &fieldData{
		path:   path,
		inMask: inMask,
	}
}

func (f fieldData) z() bool {
	return !f.value.IsValid() || reflect.ValueOf(f.value.Interface()).IsZero()
}

func (f fieldData) m() bool {
	return f.inMask
}

func (f fieldData) v() any {
	return f.value.Interface()
}

func (f fieldData) fv() protoreflect.Value {
	return f.value
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
func (store *fieldStore[T]) populate(policyFields []string, paths ...string) {
	pathLookup := make(map[string]struct{})
	for _, p := range paths {
		pathLookup[p] = struct{}{}
	}
	for _, f := range policyFields {
		_, im := pathLookup[f]
		store.populateRecursive(store.msg, store.msg.ProtoReflect().Descriptor(), im, f, "")
	}
}

// fillStore recursively ranges over the fields in the message.
func (store *fieldStore[T]) populateRecursive(message proto.Message, desc protoreflect.MessageDescriptor, inMask bool, field, traversed string) {
	if message == nil || desc == nil || field == "" || field == "." {
		return
	}
	spl := strings.Split(field, ".")
	topLevelParent := spl[0]
	var (
		fieldValue protoreflect.Value
		set        bool
	)
	existing := store.getByPath(topLevelParent)
	if existing != nil {
		if len(spl) == 1 {
			return
		}
		fieldValue = existing.fv()
		set = existing.s()
	} else {
		f := desc.Fields().ByName(protoreflect.Name(topLevelParent))
		if f == nil {
			f = desc.Fields().ByJSONName(topLevelParent)
		}
		if f == nil {
			for i := range spl {
				store.add(newUnsetFieldData(inMask, getPath(traversed, strings.Join(spl[0:i+1], "."))))
			}
			return
		}
		set = message.ProtoReflect().Has(f)
		fieldValue = message.ProtoReflect().Get(f)
	}
	if len(spl) == 1 {
		if !set {
			store.add(newUnsetFieldData(inMask, getPath(traversed, topLevelParent)))
			return
		}
		store.add(newFieldData(fieldValue, inMask, getPath(traversed, topLevelParent)))
		return
	}
	if fieldValue.Message() == nil {
		return
	}
	store.populateRecursive(
		fieldValue.Message().Interface(),
		fieldValue.Message().Descriptor(),
		inMask,
		strings.Join(spl[1:], "."),
		getPath(traversed, topLevelParent))
}

func getPath(traversed, name string) string {
	if traversed == "" {
		return name
	}
	return fmt.Sprintf("%s.%s", traversed, name)
}
