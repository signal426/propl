package policy

type TraitCalculation func(any) bool

type Trait struct {
	trait     trait
	cmpTo     any
	and       *Trait
	or        *Trait
	calculate TraitCalculation
}

func (t Trait) Calculate(v any) bool {
	return t.calculate(v)
}

func (t Trait) Trait() trait {
	return t.trait
}

func notZeroTrait() *Trait {
	return &Trait{
		trait: NotZero,
	}
}

func notEqualToTrait(v any) *Trait {
	return &Trait{
		trait: NotEq,
		cmpTo: v,
	}
}

func calculatedTrait(tc TraitCalculation) *Trait {
	return &Trait{
		trait:     Calculated,
		calculate: tc,
	}
}

type trait uint32

const (
	NotZero trait = iota
	NotEq
	Calculated
)

func (t *Trait) And(and *Trait) *Trait {
	t.and = and
	return t
}

func (t *Trait) Or(or *Trait) *Trait {
	t.or = or
	return t
}
