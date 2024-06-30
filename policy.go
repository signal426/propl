package propl

import (
	"fmt"
)

type Subject interface {
	HasTrait(t Trait) bool
	MeetsConditions(condition Condition) bool
}

type Policy struct {
	conditions Condition
	traits     *Trait
}

func NeverZeroWhen(c Condition) *Policy {
	return &Policy{
		traits:     notZeroTrait(),
		conditions: c,
	}
}

func NeverZero() *Policy {
	return &Policy{
		traits:     notZeroTrait(),
		conditions: InMessage.And(InMask),
	}
}

func Calculated(tc TraitCalculation) *Policy {
	return &Policy{
		traits:     calculatedTrait(tc),
		conditions: InMessage.And(InMask),
	}
}

func CalculatedWhen(tc TraitCalculation, c Condition) *Policy {
	return &Policy{
		traits:     calculatedTrait(tc),
		conditions: c,
	}
}

func (p *Policy) Execute(s Subject) error {
	if s.MeetsConditions(p.conditions) {
		return p.checkTraits(s, p.traits)
	}
	return nil
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
		return fmt.Errorf("does not meet policy")
	}
	// if there's an and condition, keep going
	// else, we're done
	if trait.and != nil {
		return p.checkTraits(s, trait.and)
	}
	return nil
}
