package protovalidate

import (
	"context"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type RequestValidation[T proto.Message, U any] struct {
	fieldPolicies         []*fieldPolicy[T]
	authorizationPolicies []*authzPolicy[T, U]
	rptr                  map[string]int
	rpc                   string
	paths                 PathSet
	msg                   T
	fieldMap              map[string][]protoreflect.FileDescriptor
	authorizer            U
	auth                  bool
}

func fieldDescriptors(msg proto.Message) []protoreflect.FileDescriptor {
	desc := msg.ProtoReflect().Descriptor()
	for _, f := range desc.Fields() {
	}
}

func ForAuthorizedRequest[T proto.Message, U any](rpc string, msg T, authorizer U) *RequestValidation[T, U] {
	rv := &RequestValidation[T, U]{
		rpc:                   rpc,
		fieldPolicies:         []*fieldPolicy[T]{},
		rptr:                  make(map[string]int),
		authorizer:            authorizer,
		msg:                   msg,
		authorizationPolicies: []*authzPolicy[T, U]{},
	}
	return rv
}

func ForRequest[T proto.Message](rpc string, msg T) *RequestValidation[T, NoAuth] {
	rv := &RequestValidation[T, NoAuth]{
		rpc:           rpc,
		fieldPolicies: []*fieldPolicy[T]{},
		rptr:          make(map[string]int),
		msg:           msg,
		auth:          false,
	}
	return rv
}

// MustBeNonZero validates the property exists and has a non-zero value.
// Does not evaluate field masks. Use MustBeNonZeroAndEvaluated if always required in the field
// mask or MustBeNonZeroIfEvaluated if only need to validate when in field mask.
func (r *RequestValidation[T, U]) WithFieldPolicyNeverZero(id string, value any) *RequestValidation[T, U] {
	if i, ok := r.rptr[id]; ok {
		r.fieldPolicies = append(r.fieldPolicies, newFieldPolicy[T](id, Always, value))
	} else {
		r.requirements = append(rv.requirements, nonZero(id, value))
		r.rptr[id] = len(rv.requirements) - 1
	}
}

// Do first checks the permissions of the requester (if authz requested)
// and then validates the properties of the request as directed by the caller.
// Returns a connect err.
func (rv *RequestValidation) Do(ctx context.Context) error {
	if err := rv.authz(ctx); err != nil {
		return errors.ToConnectError(ctx, err)
	}
	violations := make(map[string]error)
	for _, r := range rv.requirements {
		if err := r.validate(rv.rpc, rv.msg, rv.paths); err != nil {
			violations[r.fullPath] = err
		}
	}
	if len(violations) > 0 {
		return errors.ToConnectError(ctx, errors.NewValidationError(violations))
	}
	return nil
}
