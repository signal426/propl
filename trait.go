package propl

import (
	"fmt"
)

type TraitType uint32

const (
	NotZero TraitType = iota
	NotEqual
)

var _ Trait = (*fieldTrait)(nil)

// fieldTrait
type fieldTrait struct {
	fieldTraitType TraitType
	notEq          any
	andTrait       *fieldTrait
	orTrait        *fieldTrait
}

func (t *fieldTrait) and(and *fieldTrait) *fieldTrait {
	t.andTrait = and
	return t
}

func (t fieldTrait) NotEqVal() any {
	return t.notEq
}

func (t *fieldTrait) And() Trait {
	return t.andTrait
}

func (t fieldTrait) Type() TraitType {
	return t.fieldTraitType
}

func (t *fieldTrait) Or() Trait {
	return t.orTrait
}

func (t *fieldTrait) Valid() bool {
	return t != nil
}

func (t *fieldTrait) or(or *fieldTrait) *fieldTrait {
	t.orTrait = or
	return t
}

func (t *fieldTrait) InfractionsString() string {
	return fmt.Sprintf("it should not be zero")
}
