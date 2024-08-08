package propl

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
)

type Precheck[T proto.Message] func(ctx context.Context, msg T) error

type Subject interface {
	HasTrait(t Trait) bool
	ConditionalAction(condition Condition) Action
}

type Trait interface {
	And() Trait
	Or() Trait
	InfractionsString() string
	Type() TraitType
	Valid() bool
}

type Policy interface {
	Execute() error
	EvaluateSubjectTraits() error
}

type policy struct {
	subject    Subject
	conditions Condition
	traits     Trait
}

// Execute checks traits on the field based on the conditional action signal
// returned from the subject.
func (p *policy) Execute() error {
	switch p.subject.ConditionalAction(p.conditions) {
	case Skip:
		return nil
	case Fail:
		return fmt.Errorf("subject did not meet conditions %s", p.conditions.FlagsString())
	default:
		return p.EvaluateSubjectTraits()
	}
}

func (p *policy) EvaluateSubjectTraits() error {
	return p.checkTraits(p.traits)
}

func (p *policy) checkTraits(t Trait) error {
	if t == nil {
		return nil
	}
	if t.Valid() && !p.subject.HasTrait(t) {
		// if we have an or, keep going
		if t.Or().Valid() {
			return p.checkTraits(t.Or())
		}
		// else, we're done checking
		return errors.New(t.InfractionsString())
	}
	// if there's an and condition, keep going
	// else, we're done
	if t.And().Valid() {
		return p.checkTraits(t.And())
	}
	return nil
}

type customPolicy[T proto.Message] struct {
	arg        T
	subject    Subject
	conditions Condition
	f          func(t T) error
}

func (mp *customPolicy[T]) Execute() error {
	switch mp.subject.ConditionalAction(mp.conditions) {
	case Skip:
		return nil
	case Fail:
		return fmt.Errorf("subject did not meet conditions %s", mp.conditions.FlagsString())
	default:
		return mp.EvaluateSubjectTraits()
	}
}

func (mp *customPolicy[T]) EvaluateSubjectTraits() error {
	return mp.f(mp.arg)
}
