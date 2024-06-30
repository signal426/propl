package propl

import (
	"bytes"
)

type Condition uint32

const (
	InMessage Condition = 1 << iota
	InMask
)

func (c Condition) And(and Condition) Condition {
	c |= and
	return c
}

func (c Condition) Or(or Condition) Condition {
	c &= or
	return c
}

func (c Condition) Has(has Condition) bool {
	return c&has != 0
}

func (c Condition) FlagsString() string {
	var buffer bytes.Buffer
	if c.Has(InMessage) {
		buffer.WriteString(InMessage.String())
	}
	if c.Has(InMask) {
		if buffer.Len() > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(InMask.String())
	}
	return buffer.String()
}

type Action uint32

const (
	Check Action = iota
	Fail
	Skip
)
