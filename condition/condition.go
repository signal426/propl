package condition

type Condition uint32

const (
	InMessage Condition = iota << 1
	InMask
)

func (c Condition) And(and Condition) Condition {
	return c | and
}

func (c Condition) Has(has Condition) bool {
	return c&has != 0
}
