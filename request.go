package protopolicy

import (
	"context"

	"google.golang.org/protobuf/proto"
)

type Policy interface {
	GetBreaches(ctx context.Context) error
}

type RequestPolicy[T proto.Message] struct {
	rpc                    string
	requestMessage         T
	fieldPolicies          []*fieldPolicy[T]
	paths                  PathSet
	fieldStore             fieldStore
	validationErrHandlerFn validationErrHandlerFn
	authorizer             authorizer[T]
}

func ForRequest[T proto.Message](rpc string, msg T, opts ...RequestPolicyOption[T]) *RequestPolicy[T] {
	r := &RequestPolicy[T]{
		rpc:           rpc,
		fieldPolicies: []*fieldPolicy[T]{},
		fieldStore:    messageToFieldStore(msg, "."),
	}
	if len(opts) > 0 {
		for _, o := range opts {
			o(r)
		}
	}
	return r
}

type RequestPolicyOption[T proto.Message] func(*RequestPolicy[T])

func WithValidationHandlerFn[T proto.Message, U any](f validationErrHandlerFn) RequestPolicyOption[T] {
	return func(r *RequestPolicy[T]) {
		r.validationErrHandlerFn = f
	}
}

func (r *RequestPolicy[T]) WithAuthorizer(a authorizer[T]) RequestPolicyOption[T] {
	return func(r *RequestPolicy[T]) {
		r.authorizer = a
	}
}

func (r *RequestPolicy[T]) WithFieldPolicy(fp *fieldPolicy[T]) RequestPolicyOption[T] {
	return func(r *RequestPolicy[T]) {
		r.fieldPolicies = append(r.fieldPolicies, fp)
	}
}

func (r *RequestPolicy[T]) GetBreaches(ctx context.Context) error {
	violations := make(map[string]error)
	if r.authorizer != nil {
		if err := r.authorizer(r.requestMessage); err != nil {
			return err
		}
	}
	for _, fp := range r.fieldPolicies {
		if err := fp.check(r.rpc, r.requestMessage, r.paths); err != nil {
		}
	}
	if len(violations) > 0 {
		return r.validationErrHandlerFn(violations)
	}
	return nil
}
