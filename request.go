package protovalidate

import (
	"context"

	"google.golang.org/protobuf/proto"
)

type RequestValidation[T proto.Message, U any] struct {
	fieldPolicies          []*fieldPolicy[T]
	authorizationFn        authzFn[T, U]
	rpc                    string
	paths                  PathSet
	fieldStore             fieldStore
	authz                  bool
	validationErrHandlerFn validationErrHandlerFn
	msg                    T
}

func ForAuthorizedRequest[T proto.Message, U any](rpc string, msg T, authzCheck authzFn[T, U], opts ...RequestValidationOption[T, U]) *RequestValidation[T, U] {
	r := &RequestValidation[T, U]{
		rpc:             rpc,
		fieldPolicies:   []*fieldPolicy[T]{},
		fieldStore:      messageToFieldStore(msg, "."),
		authz:           true,
		authorizationFn: authzCheck,
	}
	if len(opts) > 0 {
		for _, o := range opts {
			o(r)
		}
	}
	return r
}

func ForRequest[T proto.Message](rpc string, msg T, opts ...UnAuthedRequestOption[T]) *RequestValidation[T, noAuth] {
	r := &RequestValidation[T, noAuth]{
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

// func (r *RequestValidation[T, U]) Authorize(ctx context.Context) error {
// 	return r.authorizationFn(r.msg, r.a)
// }

type UnAuthedRequestOption[T proto.Message] RequestValidationOption[T, noAuth]
type RequestValidationOption[T proto.Message, U any] func(*RequestValidation[T, U])

func WithValidationHandlerFn[T proto.Message, U any](f validationErrHandlerFn) RequestValidationOption[T, U] {
	return func(r *RequestValidation[T, U]) {
		r.validationErrHandlerFn = f
	}
}

func (r *RequestValidation[T, U]) WithFieldPolicies(fp ...*fieldPolicy[T]) RequestValidationOption[T, U] {
	return func(r *RequestValidation[T, U]) {
		if len(fp) > 0 {
			for _, f := range fp {
				r.fieldPolicies = append(r.fieldPolicies, f)
			}
		}
	}
}

// Do first checks the permissions of the requester (if authz requested)
// and then validates the properties of the request as directed by the caller.
// Returns a connect err.
func (r *RequestValidation[T, U]) Do(ctx context.Context) error {
	// if err := r.Authorize(ctx); err != nil {
	// 	return err
	// }
	violations := make(map[string]error)
	for _, fp := range r.fieldPolicies {
		if err := fp.check(r.rpc, r.msg, r.paths); err != nil {
		}
	}
	if len(violations) > 0 {
	}
	return nil
}
