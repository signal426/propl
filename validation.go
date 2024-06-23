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
	msg                    T
	fieldStore             fieldStore
	authorizer             U
	authz                  bool
	validationErrHandlerFn validationErrHandlerFn
}

func ForAuthorizedRequest[T proto.Message, U any](rpc string, msg T, authorizer U) *RequestValidation[T, U] {
	rv := &RequestValidation[T, U]{
		authorizer:    authorizer,
		rpc:           rpc,
		fieldPolicies: []*fieldPolicy[T]{},
		fieldStore:    MessageToFieldStore(msg, "."),
		msg:           msg,
		authz:         true,
	}
	return rv
}

func ForRequest[T proto.Message](rpc string, msg T) *RequestValidation[T, noAuth] {
	rv := &RequestValidation[T, noAuth]{
		rpc:           rpc,
		fieldPolicies: []*fieldPolicy[T]{},
		fieldStore:    MessageToFieldStore(msg, "."),
		msg:           msg,
	}
	return rv
}

func (r *RequestValidation[T, U]) Authorize(ctx context.Context) error {
	return r.authorizationFn(r.msg, r.authorizer)
}

type RequestValidationOption[T proto.Message, U any] func(*RequestValidation[T, U])

func (r *RequestValidation[T, U]) WithFieldPolicies(fp ...*fieldPolicy[T]) RequestValidationOption[T, U] {
	return func(rv *RequestValidation[T, U]) {
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
	if err := r.Authorize(ctx); err != nil {
		return err
	}
	violations := make(map[string]error)
	for _, fp := range r.fieldPolicies {
		if err := fp.validate(r.rpc, r.msg, r.paths); err != nil {
			violations[fp.fieldMeta.GetFullPath()] = err
		}
	}
	if len(violations) > 0 {
	}
	return nil
}
