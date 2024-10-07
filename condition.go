package propl

import (
	"bytes"
)

type MsgCondition uint32

const (
	InMessage MsgCondition = 1 << iota
	InMask
)

func (c MsgCondition) And(and MsgCondition) MsgCondition {
	c |= and
	return c
}

func (c MsgCondition) Or(or MsgCondition) MsgCondition {
	c &= or
	return c
}

func (c MsgCondition) Has(has MsgCondition) bool {
	return c&has != 0
}

func (c MsgCondition) FlagsString() string {
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

func AlwaysInMsg() MsgCondition {
	return InMessage.And(InMask)
}
