package propl

import (
	"errors"
	"fmt"
)

type Subject interface {
	HasTrait(t Trait) bool
	ActionFromConditions(condition Condition) Action
}

type Policy struct {
	conditions Condition
	traits     *Trait
}

func NeverZero() *Policy {
	return &Policy{
		traits:     notZeroTrait(),
		conditions: InMessage.And(InMask),
	}
}

func NeverZeroWhen(c Condition) *Policy {
	return &Policy{
		traits:     notZeroTrait(),
		conditions: c,
	}
}

func (p *Policy) And(and *Policy) *Policy {
	p.traits.And(and.traits)
	p.conditions.And(and.conditions)
	return p
}

func (p *Policy) Or(or *Policy) *Policy {
	p.traits.Or(or.traits)
	p.conditions.Or(or.conditions)
	return p
}

func Calculated(assertion string, calc func(any) bool) *Policy {
	return &Policy{
		traits: calculatedTrait(TraitCalculation{
			Assertion:   assertion,
			Calculation: calc,
		}),
		conditions: InMessage.And(InMask),
	}
}

func CalculatedWhen(assertion string, calc func(any) bool, c Condition) *Policy {
	return &Policy{
		traits: calculatedTrait(TraitCalculation{
			Assertion:   assertion,
			Calculation: calc,
		}),
		conditions: c,
	}
}

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

func (p *Policy) checkTraits(s Subject, trait *Trait) error {
	if trait == nil {
		return nil
	}
	if !s.HasTrait(*trait) {
		// if we have an or, keep going
		if trait.or != nil {
			return p.checkTraits(s, trait.or)
		}
		// else, we're done checking
		return errors.New(trait.ViolationString())
	}
	// if there's an and condition, keep going
	// else, we're done
	if trait.and != nil {
		return p.checkTraits(s, trait.and)
	}
	return nil
}
