package propl

import (
	"context"

	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"
)

// Propl is an aggregation of policies on some proto message.
type Propl[T proto.Message] struct {
	policies                map[string]*Policy
	fieldStore              *fieldStore[T]
	fieldInfractionsHandler FieldInfractionsHandler
	precheck                Precheck[T]
}

// For creates a new policy aggregate for the specified message that can be built upon using the
// builder methods.
func For[T proto.Message](msg T, paths ...string) *Propl[T] {
	r := &Propl[T]{
		fieldStore: newFieldStore(msg),
		policies:   make(map[string]*Policy),
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
	r.policies[path] = &Policy{
		conditions: conditions,
		traits:     traits,
	}
	return r
}

// WithFieldPolicy adds a field policy for the request. Accepts a "." delimited location to the
// field to which the policy applies.
// Duplicate path entries results in the last policy set to be the one applied.
func (r *Propl[T]) NeverZero(path string) *Propl[T] {
	fp := &Policy{
		conditions: InMask.And(InMessage),
		traits: &fieldTrait{
			fieldTraitType: NotZero,
		},
	}
	r.policies[path] = fp
	return r
}

func (r *Propl[T]) NeverZeroWhen(path string, conditions Condition) *Propl[T] {
	fp := &Policy{
		conditions: conditions,
		traits: &fieldTrait{
			fieldTraitType: NotZero,
		},
	}
	r.policies[path] = fp
	return r
}

func (r *Propl[T]) NeverEq(path string, v any) *Propl[T] {
	fp := &Policy{
		conditions: InMask.And(InMessage),
		traits: &fieldTrait{
			fieldTraitType: NotEqual,
			notEq:          v,
		},
	}
	r.policies[path] = fp
	return r
}

func (r *Propl[T]) NeverEqWhen(path string, conditions Condition, v any) *Propl[T] {
	fp := &Policy{
		conditions: conditions,
		traits: &fieldTrait{
			fieldTraitType: NotEqual,
			notEq:          v,
		},
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
	// populate the store based on the validation fields
	// and the message
	r.fieldStore.populate(maps.Keys(r.policies))
	// ensure some handler is set
	r.ensureFieldInfractionsHandler()
	finfractions := make(map[string]error)
	for id, p := range r.policies {
		if err := p.Execute(r.fieldStore.getByPath(id)); err != nil {
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
