package protopolicy

import (
	"context"

	"github.com/signal426/protopolicy/policy"
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
	r := &RequestPolicy[T]{
		rpc:            rpc,
		requestMessage: msg,
		fieldPolicies:  []*fieldPolicy{},
		fieldStore:     messageToFieldStore(msg, ".", paths...),
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

func (r *RequestPolicy[T]) WithFieldPolicy(path string, policy *policy.Policy) *RequestPolicy[T] {
	r.fieldPolicies = append(r.fieldPolicies, &fieldPolicy{
		policy: policy,
		field:  r.fieldStore.getByPath(path),
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
