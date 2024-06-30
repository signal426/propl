package propl

import "fmt"

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

type TraitCalculation struct {
	Calculation func(any) bool
	ItShouldNot string
}

// Trait
type Trait struct {
	trait     trait
	and       *Trait
	or        *Trait
	calculate TraitCalculation
}

func (t Trait) Calculate(v any) bool {
	return t.calculate.Calculation(v)
}

func (t Trait) Trait() trait {
	return t.trait
}

func (t Trait) ViolationString() string {
	if t.trait == calculated {
		return fmt.Sprintf("it should not %s", t.calculate.ItShouldNot)
	}
	return fmt.Sprintf("it should not be zero")
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
