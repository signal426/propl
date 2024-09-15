package propl

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type UhOhType uint32

const (
	NA UhOhType = 1 << iota
	Field
	Request
)

func FieldUhOh(field string, err error) UhOh {
	return UhOh{
		Type:  Field,
		Err:   err,
		Field: field,
	}
}

func RequestUhOh(err error) UhOh {
	return UhOh{
		Type: Request,
		Err:  err,
	}
}

type UhOh struct {
	Type  UhOhType
	Field string
	Err   error
}

type ByType []UhOh

func (u ByType) Len() int {
	return len(u)
}

func (u ByType) Less(i, j int) bool {
	return u[i].Type < u[j].Type
}

func (u ByType) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}

type UhOhHandler interface {
	SpaghettiOs(uhohs []UhOh) error
}

var _ UhOhHandler = (*defaultUhOhHandler)(nil)

type defaultUhOhHandler struct{}

func newDefaultUhOhHandler() *defaultUhOhHandler {
	return &defaultUhOhHandler{}
}

// Process implements ErrResultHandler.
func (*defaultUhOhHandler) SpaghettiOs(uhohs []UhOh) error {
	var (
		buffer         bytes.Buffer
		sectionWritten string
	)
	// sort the uhohs by type so that we can process the sections easier
	sort.Sort(ByType(uhohs))
	for _, v := range uhohs {
		if sectionWritten == "" || sectionWritten != v.Type.String() {
			// if we are starting a new sction, end the current one
			if sectionWritten != "" {
				buffer.WriteString("]\n")
			}
			buffer.WriteString(fmt.Sprintf("%s.issues: [\n", strings.ToLower(v.Type.String())))
			sectionWritten = v.Type.String()
		}
		var lineitem string
		if v.Field == "" {
			lineitem = fmt.Sprintf("%s\\n\n", v.Err.Error())
		} else {
			lineitem = fmt.Sprintf("%s: %s\\n\n", v.Field, v.Err.Error())
		}
		buffer.WriteString(lineitem)
	}
	buffer.WriteString("]\n")
	return errors.New(buffer.String())
}

// parseFieldNameFromPath parses the target field's name from a "." delimited path.
// returns the parent path and the field's name respectively.
func parseFieldNameFromPath(path string) (string, string) {
	sp := strings.Split(path, ".")
	var parsedName, parentPath string
	if len(sp) > 1 {
		parsedName = sp[len(sp)-1]
		parentPath = strings.Join(sp[:len(sp)-1], ".")
	} else {
		parsedName = sp[0]
	}
	return parentPath, parsedName
}
