package protopolicy

type Condition uint32

const (
	InMessage Condition = iota << 1
	InMask
)

func (c Condition) And(and Condition) Condition {
	return c | and
}

func (c Condition) Or(or Condition) Condition {
	return c & or
}

func (c Condition) Has(has Condition) bool {
	return c&has != 0
}
