package propl

import (
	"context"

	"google.golang.org/protobuf/proto"
)

// fieldPolicy ties field data to a policy
type fieldPolicy struct {
	id     string
	field  *fieldData
	policy *Policy
}

func (r *fieldPolicy) check() error {
	return r.policy.Execute(r.field)
}

type RequestPolicy[T proto.Message] struct {
	rpc                     string
	requestMessage          T
	fieldPolicies           []*fieldPolicy
	fieldStore              fieldStore
	fptr                    map[string]int
	fieldInfractionsHandler FieldInfractionsHandler
	precheck                Precheck[T]
}

func ForRequest[T proto.Message](rpc string, msg T, paths ...string) *RequestPolicy[T] {
	fieldStore := newFieldStore()
	fieldStore.fill(msg, paths...)
	r := &RequestPolicy[T]{
		rpc:            rpc,
		requestMessage: msg,
		fieldPolicies:  []*fieldPolicy{},
		fieldStore:     fieldStore,
		fptr:           make(map[string]int),
	}
	return r
}

// WithInfractionsHandler specify how to handle the infractions map (map[string]error) if there are any
func (r *RequestPolicy[T]) WithFieldInfractionsHandler(f FieldInfractionsHandler) *RequestPolicy[T] {
	r.fieldInfractionsHandler = f
	return r
}

// WithPrecheckPolicy executes before field policies are evaluated. The check exits and does not evaluate
// fields if the precheck returns an error.
func (r *RequestPolicy[T]) WithPrecheckPolicy(p Precheck[T]) *RequestPolicy[T] {
	r.precheck = p
	return r
}

// WithFieldPolicy adds a field policy for the request. Accepts a "." delimited location to the
// field to which the policy applies.
// Duplicate path entries results in the last policy set to be the one applied.
func (r *RequestPolicy[T]) WithFieldPolicy(path string, policy *Policy) *RequestPolicy[T] {
	fieldData := r.fieldStore.getByPath(path)
	if fieldData == nil {
		fieldData = newUnsetFieldData(path, false)
		r.fieldStore.add(fieldData)
	}
	fieldPolicy := &fieldPolicy{
		policy: policy,
		field:  fieldData,
		id:     path,
	}
	// overwrite if we already have a field policy for this path
	if idx, ok := r.fptr[fieldPolicy.id]; ok {
		r.fieldPolicies[idx] = fieldPolicy
	} else {
		r.fieldPolicies = append(r.fieldPolicies, fieldPolicy)
		r.fptr[fieldPolicy.id] = len(r.fieldPolicies) - 1
	}
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
	if r.precheck != nil {
		if err := r.precheck(r.requestMessage); err != nil {
			return err
		}
	}
	r.setFieldInfractionsHandler()
	finfractions := make(map[string]error)
	for _, fp := range r.fieldPolicies {
		if err := fp.check(); err != nil {
			finfractions[fp.id] = err
		}
	}
	if len(finfractions) > 0 {
		return r.fieldInfractionsHandler(finfractions)
	}
	return nil
}

func (r *RequestPolicy[T]) setFieldInfractionsHandler() {
	if r.fieldInfractionsHandler == nil {
		r.fieldInfractionsHandler = defaultFieldInfractionsHandler
	}
}
