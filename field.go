package propl

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type validationErrHandlerFn func(errs map[string]error) error

func defaultValidationErrHandlerFn(errs map[string]error) error {
	var buffer bytes.Buffer
	buffer.WriteString("field violations: [")
	for k, v := range errs {
		buffer.WriteString(fmt.Sprintf("%s: %s,\n", k, v.Error()))
	}
	buffer.WriteString("]")
	return errors.New(buffer.String())
}

type fieldPolicy struct {
	id     string
	field  *fieldData
	policy *Policy
}

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
