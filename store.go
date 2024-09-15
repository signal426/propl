package propl

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var _ PolicySubjectStore = (*fieldStore)(nil)

type fieldStore struct {
	msg            proto.Message
	maskPathLookup map[string]struct{}
	store          map[string]*fieldData
}

// Subject implements PolicySubjectStore.
func (s *fieldStore) Subject() proto.Message {
	return s.msg
}

// GetByID implements PolicySubjectStore.
func (s *fieldStore) GetByID(_ context.Context, id string) PolicySubject {
	return s.dataAtPath(id)
}

// ProcessByID implements PolicySubjectStore.
func (s *fieldStore) ProcessByID(ctx context.Context, id string) PolicySubject {
	s.processPathRecursive(s.msg, s.msg.ProtoReflect().Descriptor(), s.isFieldInMask(id), id, "")
	return s.dataAtPath(id)
}

type initFieldStoreOption func(*fieldStore)

func withPaths(paths ...string) initFieldStoreOption {
	return func(fs *fieldStore) {
		pathLookup := make(map[string]struct{})
		for _, p := range paths {
			pathLookup[p] = struct{}{}
		}
		fs.maskPathLookup = pathLookup
	}
}

func initFieldStore(msg proto.Message, options ...initFieldStoreOption) *fieldStore {
	s := &fieldStore{
		msg:   msg,
		store: make(map[string]*fieldData),
	}
	if len(options) > 0 {
		for _, o := range options {
			o(s)
		}
	}
	return s
}

func (f fieldStore) message() proto.Message {
	return f.msg
}

func (f fieldStore) isFieldInMask(p string) bool {
	if _, im := f.maskPathLookup[p]; im {
		return im
	}
	if s := strings.Split(p, "."); len(s) > 0 {
		if _, im := f.maskPathLookup[s[len(s)-1]]; im {
			return im
		}
	}
	return false
}

func (f fieldStore) empty() bool {
	return f.store == nil || len(f.store) == 0
}

func (f *fieldStore) add(fd *fieldData) {
	f.store[fd.p()] = fd
}

func (f fieldStore) dataAtPath(p string) *fieldData {
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

var _ PolicySubject = (*fieldData)(nil)

type fieldData struct {
	path   string
	inMask bool
	set    bool
	value  protoreflect.Value
}

// Evaluatable implements PolicySubject.
func (f *fieldData) Evaluatable(conditions Condition) bool {
	if !conditions.Has(InMessage) && conditions.Has(InMask) && !f.m() {
		return false
	}
	return true
}

// MeetsConditions implements PolicySubject.
func (f *fieldData) MeetsConditions(conditions Condition) bool {
	if !f.Evaluatable(conditions) {
		return false
	}
	if !f.s() && conditions.Has(InMessage) {
		return false
	}
	if !f.s() && conditions.Has(InMask) && f.m() {
		return false
	}
	return true
}

// ID implements PolicySubject.
func (f *fieldData) ID() string {
	return f.p()
}

// HasTrait implements policy.Subject.
func (f *fieldData) HasTrait(t Trait) bool {
	return t.Type() == NotZero && f.z()
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

func (store *fieldStore) processPathRecursive(message proto.Message, desc protoreflect.MessageDescriptor, inMask bool, field, traversed string) {
	if message == nil || desc == nil || field == "" || field == "." {
		return
	}
	spl := strings.Split(field, ".")
	topLevelParent := spl[0]
	var (
		fieldValue protoreflect.Value
		set        bool
	)
	existing := store.dataAtPath(topLevelParent)
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
	store.processPathRecursive(
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
