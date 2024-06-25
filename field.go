package protovalidate

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
)

type Trait uint32

const (
	NotZero Trait = iota << 1
	NotEq
	Custom
)

type Condition uint32

const (
	InMessage Condition = iota << 1
	InMask
)

func And[T uint32](this, and T) T {
	return this | and
}

func Has[T uint32](this, has T) bool {
	return this&has != 0
}

type validationErrHandlerFn func(errs map[string]error) error

func defaultValidationErrHandlerFn(errs map[string]error) error {
	var buffer bytes.Buffer
	buffer.WriteString("field violations: [")
	for k, v := range errs {
		buffer.WriteString(fmt.Sprintf("%s: %s,\n", k, v.Error()))
	}
	buffer.WriteString("]")
	return errors.New(buffer.String())
}

type policy[T proto.Message] func(t T) error

type fieldMeta struct {
	id         string
	value      any
	parentPath string
	fullPath   string
}

func (f *fieldMeta) GetID() string {
	if f == nil {
		return ""
	}
	return f.id
}

func (f *fieldMeta) GetValue() any {
	if f == nil {
		return nil
	}
	return f.value
}

func (f *fieldMeta) GetParentPath() string {
	if f == nil {
		return ""
	}
	return f.parentPath
}

func (f *fieldMeta) GetFullPath() string {
	if f == nil {
		return ""
	}
	return f.fullPath
}

type fieldMetaOption func(*fieldMeta)

func fieldMetaWithValue(v any) fieldMetaOption {
	return func(fm *fieldMeta) {
		fm.value = v
	}
}

func newFieldMeta(id string, opts ...fieldMetaOption) *fieldMeta {
	parsedID, parentPath := parseID(id)
	fm := &fieldMeta{
		id:         parsedID,
		parentPath: parentPath,
		fullPath:   id,
	}
	if len(opts) > 0 {
		for _, o := range opts {
			o(fm)
		}
	}
	return fm
}

type fieldPolicy[T proto.Message] struct {
	meta       *fieldMeta
	notEq      any
	conditions Condition
	traits     Trait
	policy     policy[T]
}

func parseID(id string) (string, string) {
	sp := strings.Split(id, ".")
	var parsedID, parentPath string
	if len(sp) > 1 {
		parsedID = sp[len(sp)-1]
		parentPath = strings.Join(sp[:len(sp)-1], ".")
	} else {
		parsedID = sp[0]
	}
	return parsedID, parentPath
}

func NewFieldPolicy(id string, cond PolicyCondition, value any, notEq any) *fieldPolicy[T] {
	return &fieldPolicy{
		conditions: cond,
		meta:       newFieldMeta(id, fieldMetaWithValue(value)),
		notEq:      notEq,
	}
}

func newCustomFieldPolicy[T proto.Message](id string, policy policy[T]) *fieldPolicy[T] {
	return &fieldPolicy[T]{
		conditions: Custom,
		meta:       newFieldMeta(id),
		policy:     policy,
	}
}

// inMask returns two booleans -- first to indicate if the paths were
// queryable and second to indicate if the field is in the collection of paths.
func (r *fieldPolicy[T]) inMask(paths PathSet) (bool, bool) {
	if paths.Empty() {
		return false, false
	}
	return true, paths.Has(r.meta.GetID())
}

func (r *fieldPolicy[T]) check(rpc string, msg T, paths PathSet) error {
	return nil
}
