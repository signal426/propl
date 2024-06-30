package propl

import (
	"context"

	"google.golang.org/protobuf/proto"
)

type RequestPolicy[T proto.Message] struct {
	rpc                    string
	requestMessage         T
	fieldPolicies          []*fieldPolicy
	fieldStore             fieldStore
	validationErrHandlerFn validationErrHandlerFn
	authorizer             authorizer[T]
}

func ForRequest[T proto.Message](rpc string, msg T, paths ...string) *RequestPolicy[T] {
	fieldStore := newFieldStore()
	fieldStore.fill(msg, paths...)
	r := &RequestPolicy[T]{
		rpc:            rpc,
		requestMessage: msg,
		fieldPolicies:  []*fieldPolicy{},
		fieldStore:     fieldStore,
	}
	return r
}

func (r *RequestPolicy[T]) WithValidationHandlerFn(f validationErrHandlerFn) *RequestPolicy[T] {
	r.validationErrHandlerFn = f
	return r
}

func (r *RequestPolicy[T]) WithAuthorizer(a authorizer[T]) *RequestPolicy[T] {
	r.authorizer = a
	return r
}

func (r *RequestPolicy[T]) WithFieldPolicy(path string, policy *Policy) *RequestPolicy[T] {
	fieldData := r.fieldStore.getByPath(path)
	if fieldData == nil {
		fieldData = newUnsetFieldData(path, false)
		r.fieldStore.add(fieldData)
	}
	r.fieldPolicies = append(r.fieldPolicies, &fieldPolicy{
		policy: policy,
		field:  fieldData,
		id:     path,
	})
	return r
}

func (r *RequestPolicy[T]) GetViolations(ctx context.Context) error {
	if r.validationErrHandlerFn == nil {
		r.validationErrHandlerFn = defaultValidationErrHandlerFn
	}
	violations := make(map[string]error)
	if r.authorizer != nil {
		if err := r.authorizer(r.requestMessage); err != nil {
			return err
		}
	}
	for _, fp := range r.fieldPolicies {
		if err := fp.check(); err != nil {
			violations[fp.id] = err
		}
	}
	if len(violations) > 0 {
		return r.validationErrHandlerFn(violations)
	}
	return nil
}
