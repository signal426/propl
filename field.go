package propl

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type InfractionsHandler func(errs map[string]error) error

// defaultValidationErrHandlerFn if no ViolationsHandler specified
func defaultInfractionsHandler(errs map[string]error) error {
	var buffer bytes.Buffer
	buffer.WriteString("field infractions: [\n")
	for k, v := range errs {
		buffer.WriteString(fmt.Sprintf("%s: %s\\n\n", k, v.Error()))
	}
	buffer.WriteString("]\n")
	return errors.New(buffer.String())
}

// fieldPolicy ties field data to a policy
type fieldPolicy struct {
	id     string
	field  *fieldData
	policy *Policy
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

func (r *fieldPolicy) check() error {
	return r.policy.Execute(r.field)
}
