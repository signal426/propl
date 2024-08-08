package propl

import (
	"fmt"
)

type TraitType uint32

const (
	NotZero TraitType = iota
	NotEqual
)

var _ Trait = (*trait)(nil)

// trait
type trait struct {
	traitType TraitType
	andTrait  *trait
	orTrait   *trait
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

func (t *trait) InfractionsString() string {
	return fmt.Sprintf("it should not be zero")
}
