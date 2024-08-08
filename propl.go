package propl

import (
	"context"

	"google.golang.org/protobuf/proto"
)

// Propl is an aggregation of policies on some proto message.
type Propl[T proto.Message] struct {
	policies                map[string]Policy
	fieldStore              *fieldStore[T]
	fieldInfractionsHandler FieldInfractionsHandler
	precheck                Precheck[T]
}

// For creates a new policy aggregate for the specified message that can be built upon using the
// builder methods.
func For[T proto.Message](msg T, paths ...string) *Propl[T] {
	r := &Propl[T]{
		fieldStore: newFieldStore(msg, paths...),
		policies:   make(map[string]Policy),
	}
	return r
}

// WithInfractionsHandler specify how to handle the infractions map (map[string]error) if there are any
func (r *Propl[T]) WithFieldInfractionsHandler(f FieldInfractionsHandler) *Propl[T] {
	r.fieldInfractionsHandler = f
	return r
}

// WithPrecheckPolicy executes before field policies are evaluated. The check exits and does not evaluate
// fields if the precheck returns an error.
func (r *Propl[T]) WithPrecheckPolicy(p Precheck[T]) *Propl[T] {
	r.precheck = p
	return r
}

func (r *Propl[T]) FieldPolicy(path string, traits Trait, conditions Condition) *Propl[T] {
	r.policies[path] = &policy{
		conditions: conditions,
		traits:     traits,
	}
	return r
}

// NeverZero validates that the field at the provided path
// is always (in body or mask) non-zero
func (r *Propl[T]) NeverZero(path string) *Propl[T] {
	fp := &policy{
		subject:    r.fieldStore.loadFieldsFromPath(path).getByPath(path),
		conditions: InMask.And(InMessage),
		traits: &trait{
			traitType: NotZero,
		},
	}
	r.policies[path] = fp
	return r
}

// NeverZeroWhen validates that the field at the provided location is
// not zero under the provided conditions (e.g. in a field mask)
func (r *Propl[T]) NeverZeroWhen(path string, conditions Condition) *Propl[T] {
	fp := &policy{
		subject:    r.fieldStore.loadFieldsFromPath(path).getByPath(path),
		conditions: conditions,
		traits: &trait{
			traitType: NotZero,
		},
	}
	r.policies[path] = fp
	return r
}

// CustomEval asserts the field is always present and set before running
// a user-provided function that receives the entire message as an arg
func (r *Propl[T]) CustomEval(path string, c func(t T) error) *Propl[T] {
	fp := &customPolicy[T]{
		conditions: InMask.And(InMessage),
		arg:        r.fieldStore.message(),
		subject:    r.fieldStore.loadFieldsFromPath(path).getByPath(path),
		f:          c,
	}
	r.policies[path] = fp
	return r
}

// CustomEvalWhen runs a custom eval function that receives the entire message as an arg
// when the field at the specified location meets the specified conditions
func (r *Propl[T]) CustomEvalWhen(path string, conditions Condition, c func(t T) error) *Propl[T] {
	fp := &customPolicy[T]{
		conditions: conditions,
		arg:        r.fieldStore.message(),
		subject:    r.fieldStore.loadFieldsFromPath(path).getByPath(path),
		f:          c,
	}
	r.policies[path] = fp
	return r
}

// E shorthand for Evaluate
func (r *Propl[T]) E(ctx context.Context) error {
	return r.Evaluate(ctx)
}

// Evaluate checks each declared policy and returns an error describing
// each infraction. If a precheck is specified and returns an error, this exits
// and field policies are not evaluated.
//
// To use your own infractionsHandler, specify a handler using WithInfractionsHandler.
func (r *Propl[T]) Evaluate(ctx context.Context) error {
	if r.precheck != nil {
		if err := r.precheck(ctx, r.fieldStore.message()); err != nil {
			return err
		}
	}
	// ensure some handler is set
	r.ensureFieldInfractionsHandler()
	finfractions := make(map[string]error)
	for id, p := range r.policies {
		if err := p.Execute(); err != nil {
			finfractions[id] = err
		}
	}
	if len(finfractions) > 0 {
		return r.fieldInfractionsHandler(finfractions)
	}
	return nil
}

func (r *Propl[T]) ensureFieldInfractionsHandler() {
	if r.fieldInfractionsHandler == nil {
		r.fieldInfractionsHandler = defaultFieldInfractionsHandler
	}
}
