package propl

type TraitType uint32

const (
	NotZero TraitType = iota
	NotEqual
)

var _ Trait = (*trait)(nil)

// trait is a feature of a subject
type trait struct {
	// the trait type
	traitType TraitType
	// the trait that this trait is composed with
	andTrait *trait
	// the trait that this trait is composed with
	orTrait *trait
}

func (t *trait) and(and *trait) *trait {
	t.andTrait = and
	return t
}

func (t *trait) And() Trait {
	return t.andTrait
}

func (t trait) Type() TraitType {
	return t.traitType
}

func (t *trait) Or() Trait {
	return t.orTrait
}

func (t *trait) Valid() bool {
	return t != nil
}

func (t *trait) or(or *trait) *trait {
	t.orTrait = or
	return t
}

func (t *trait) UhOhString() string {
	if t.Type() == NotZero {
		return "it should not be zero"
	}
	return "it should not be equal to the supplied value"
}
