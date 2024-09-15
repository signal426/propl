package propl

import (
	"context"

	"google.golang.org/protobuf/proto"
)

// Some function triggered by the result of an evaluation whether it be
// a policy or a global evaluation
type TriggeredEvaluation func(ctx context.Context, msg proto.Message) error

// a policy subject is a subject that gets evaluated to see:
// 1. what action is configured to occurr if a certain condition is met
// 2. what traits it has if the conditional action results in a trait eval
type PolicySubject interface {
	// some identifier for the subject
	ID() string
	// check for whether or not a policy subject holds a trait
	HasTrait(t Trait) bool
	// evaluatable reports whether or not the subject is in a state
	// that resolves to some action. for example, if a policy that ensures that
	// a subject has a non-zero value but the evaluation condition is a in the
	// update mask (and the field was not supplied in the mask), the field is not
	// evaluatable
	Evaluatable(conditions Condition) bool
	// reports whether or not the subject meets the supplied conditions. Anything that
	// returns false for Evaluatable will return false for MeetsConditions
	MeetsConditions(conditions Condition) bool
}

// a trait is an attribute of a policy subject that must
// be true if the policy is a trait evaluation
type Trait interface {
	// another trait that must exist with this trait
	And() Trait
	// another trait that must exist if this trait does not
	Or() Trait
	// some error string describing the validation error
	UhOhString() string
	// the trait type
	Type() TraitType
	// state check to report the validity of trait
	Valid() bool
}

// a policy executes its configured logic against the specified subject
type Policy interface {
	EvaluateSubject(ctx context.Context, subject PolicySubject) error
}

// a way to store and retrieve policy subjects as well
// as holds a reference to the primary subject
type PolicySubjectStore interface {
	// store the id
	ProcessByID(ctx context.Context, id string) PolicySubject
	// get the policy subject by some id
	GetByID(ctx context.Context, id string) PolicySubject
	// subject is the message that the store was hydrated from
	Subject() proto.Message
}
