package propl

type trait uint32

const (
	notZero trait = iota
	calculated
)

func (t *Trait) And(and *Trait) *Trait {
	t.and = and
	return t
}

func (t *Trait) Or(or *Trait) *Trait {
	t.or = or
	return t
}

type TraitCalculation func(any) bool

type Trait struct {
	trait     trait
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
		trait: notZero,
	}
}

func calculatedTrait(tc TraitCalculation) *Trait {
	return &Trait{
		trait:     calculated,
		calculate: tc,
	}
}
