package policy

import (
	"fmt"
)

type Subject interface {
	HasTrait(t Trait) bool
	MeetsConditions(condition Condition) bool
}

type Policy struct {
	condition Condition
	traits    *Trait
}

func NeverZeroWhen(c Condition) *Policy {
	return &Policy{
		traits:    notZeroTrait(),
		condition: c,
	}
}

func NeverZero() *Policy {
	return &Policy{
		traits:    notZeroTrait(),
		condition: InMessage.And(InMask),
	}
}

func NeverEqual(v any) *Policy {
	return &Policy{
		traits:    notEqualToTrait(v),
		condition: InMessage.And(InMask),
	}
}

func NeverEqualWhen(v any, c Condition) *Policy {
	return &Policy{
		traits:    notEqualToTrait(v),
		condition: c,
	}
}

func CalculatedTrait(tc TraitCalculation) *Policy {
	return &Policy{
		traits:    calculatedTrait(tc),
		condition: InMessage.And(InMask),
	}
}

func CalculatedTraitWhen(tc TraitCalculation, c Condition) *Policy {
	return &Policy{
		traits:    calculatedTrait(tc),
		condition: c,
	}
}

func (p *Policy) Execute(s Subject) error {
	if s.MeetsConditions(p.condition) && p.traits != nil {
		return p.checkTraits(s, p.traits, nil)
	}
	return fmt.Errorf("does not meet conditions")
}

func (p *Policy) checkTraits(s Subject, trait *Trait, prev *Trait) error {
	if trait == nil {
		return nil
	}
	if !s.HasTrait(*trait) {
		if trait.or != nil {
			return p.checkTraits(s, trait.or, trait)
		}
		return fmt.Errorf("does not meet policy")
	}
	if trait.and != nil {
		return p.checkTraits(s, trait.and, trait)
	}
	return nil
}
