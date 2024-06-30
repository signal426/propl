package propl

import (
	"context"

	"google.golang.org/protobuf/proto"
)

type RequestPolicy[T proto.Message] struct {
	rpc                string
	requestMessage     T
	fieldPolicies      []*fieldPolicy
	fieldStore         fieldStore
	infractionsHandler InfractionsHandler
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

// WithInfractionsHandler specify how to handle the infractions map if there are any
// (map[string]error)
func (r *RequestPolicy[T]) WithInfractionsHandler(f InfractionsHandler) *RequestPolicy[T] {
	r.infractionsHandler = f
	return r
}

// WithFieldPolicy adds a field policy for the request. Accepts a "." delimited location to the
// field to which the policy applies.
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

// E shorthand for Evaluate
func (r *RequestPolicy[T]) E(ctx context.Context) error {
	return r.Evaluate(ctx)
}

// Evaluate checks each declared policy and returns an error describing
// each infraction.
// To use your own infractionsHandler, specify a handler using WithInfractionsHandler.
func (r *RequestPolicy[T]) Evaluate(ctx context.Context) error {
	if r.infractionsHandler == nil {
		r.infractionsHandler = defaultInfractionsHandler
	}
	infractions := make(map[string]error)
	for _, fp := range r.fieldPolicies {
		if err := fp.check(); err != nil {
			infractions[fp.id] = err
		}
	}
	if len(infractions) > 0 {
		return r.infractionsHandler(infractions)
	}
	return nil
}
