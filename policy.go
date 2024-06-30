package propl

import (
	"errors"
	"fmt"
)

type Subject interface {
	HasTrait(t trait) bool
	ActionFromConditions(condition Condition) Action
}

type Policy struct {
	conditions Condition
	traits     *trait
}

// NeverZero triggers a violation if
// the field is zero.
func NeverZero() *Policy {
	return &Policy{
		traits:     notZeroTrait(),
		conditions: InMessage.And(InMask),
	}
}

// NeverZeroWhen triggers a violation if
// the field has the specified condition(s)
func NeverZeroWhen(c Condition) *Policy {
	return &Policy{
		traits:     notZeroTrait(),
		conditions: c,
	}
}

// And chains a policy to another policy.
// If one policy fails, the chain fails.
func (p *Policy) And(and *Policy) *Policy {
	p.traits.and(and.traits)
	p.conditions.And(and.conditions)
	return p
}

// Calculated runs the specified function if field is set.
func Calculated(assertion string, calc func(any) bool) *Policy {
	return &Policy{
		traits: calculatedTrait(traitCalculation{
			assertion:   assertion,
			calculation: calc,
		}),
		conditions: InMessage.And(InMask),
	}
}

// CalculatedWhen runs the specified function when the conditions apply.
func CalculatedWhen(assertion string, calc func(any) bool, c Condition) *Policy {
	return &Policy{
		traits: calculatedTrait(traitCalculation{
			assertion:   assertion,
			calculation: calc,
		}),
		conditions: c,
	}
}

// Execute checks traits on the field based on the action signal
// specified from the subject.
func (p *Policy) Execute(s Subject) error {
	switch s.ActionFromConditions(p.conditions) {
	case Skip:
		return nil
	case Fail:
		return fmt.Errorf("did not meet conditions %s", p.conditions.FlagsString())
	default:
		return p.checkTraits(s, p.traits)
	}
}

func (p *Policy) checkTraits(s Subject, trait *trait) error {
	if trait == nil {
		return nil
	}
	if !s.HasTrait(*trait) {
		// if we have an or, keep going
		if trait.orTrait != nil {
			return p.checkTraits(s, trait.orTrait)
		}
		// else, we're done checking
		return errors.New(trait.ViolationString())
	}
	// if there's an and condition, keep going
	// else, we're done
	if trait.andTrait != nil {
		return p.checkTraits(s, trait.andTrait)
	}
	return nil
}
