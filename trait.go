package propl

import "fmt"

type traitType uint32

const (
	notZero traitType = iota
	calculated
)

func (t *trait) and(and *trait) *trait {
	t.andTrait = and
	return t
}

func (t *trait) or(or *trait) *trait {
	t.orTrait = or
	return t
}

type traitCalculation struct {
	calculation func(any) bool
	assertion   string
}

// trait
type trait struct {
	traitType   traitType
	andTrait    *trait
	orTrait     *trait
	calculation traitCalculation
}

func (t trait) calculate(v any) bool {
	return t.calculation.calculation(v)
}

func (t trait) infractionString() string {
	if t.traitType == calculated {
		return fmt.Sprintf("%s", t.calculation.assertion)
	}
	return fmt.Sprintf("it should not be zero")
}

func notZeroTrait() *trait {
	return &trait{
		traitType: notZero,
	}
}

func calculatedTrait(tc traitCalculation) *trait {
	return &trait{
		traitType:   calculated,
		calculation: tc,
	}
}
