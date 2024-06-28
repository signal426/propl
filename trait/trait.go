package trait

type Trait struct {
	trait trait
	cmpTo any
	and   *Trait
}

func NotEqualTo(v any) *Trait {
	return &Trait{
		trait: notEq,
		cmpTo: v,
	}
}

type trait uint32

const (
	notZero trait = iota
	notEq
	custom
)

func (t *Trait) And(and *Trait) *Trait {
	t.and = and
	return t
}
