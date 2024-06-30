package propl

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type FieldInfractionsHandler func(errs map[string]error) error

// defaultFieldInfractionsHandler if no FieldInfractionsHandler specified
func defaultFieldInfractionsHandler(errs map[string]error) error {
	var buffer bytes.Buffer
	buffer.WriteString("field infractions: [\n")
	for k, v := range errs {
		buffer.WriteString(fmt.Sprintf("%s: %s\\n\n", k, v.Error()))
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
