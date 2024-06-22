package protovalidate

import (
	"reflect"
	"strings"

	"google.golang.org/protobuf/proto"
)

type ValidationErr struct {
	Key     string
	Details string
}

type validationFn[T proto.Message] func(t T) error

type RequirementCondition uint32

const (
	Always RequirementCondition = 1 << iota
	InMask
	NotEqual
	Custom
)

func (r RequirementCondition) Add(toAdd RequirementCondition) RequirementCondition {
	return r | toAdd
}

func (r RequirementCondition) Has(has RequirementCondition) bool {
	return r&has != 0
}

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
	fieldMeta    *fieldMeta
	conditions   RequirementCondition
	validationFn validationFn[T]
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

func newFieldPolicy[T proto.Message](id string, cond RequirementCondition, value any, validationFn validationFn[T]) *fieldPolicy[T] {
	return &fieldPolicy[T]{
		conditions:   cond,
		validationFn: validationFn,
		fieldMeta:    newFieldMeta(id, fieldMetaWithValue(value)),
	}
}

// inMask returns two booleans -- first to indicate if the paths were
// queryable and second to indicate if the field is in the collection of paths.
func (r *fieldPolicy[T]) inMask(paths PathSet) (bool, bool) {
	if paths.Empty() {
		return false, false
	}
	return true, paths.Has(r.fieldMeta.GetID())
}

// todo: maybe remove if get field already does this
func (r *fieldPolicy[T]) HasValue() bool {
	return r.fieldMeta.GetValue() != nil && !reflect.ValueOf(r.fieldMeta.GetValue()).IsZero()
}

func (r *fieldPolicy[T]) validate(rpc string, msg proto.Message, paths PathSet) error {
	return nil
}
