package protopolicy

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/signal426/protopolicy/policy"
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
	policy *policy.Policy
}

func parseID(id string) (string, string) {
	sp := strings.Split(id, ".")
	var parsedID, parentPath string
	if len(sp) > 1 {
		parsedID = sp[len(sp)-1]
		parentPath = strings.Join(sp[:len(sp)-1], ".")
	} else {
		parsedID = sp[0]
	}
	return parsedID, parentPath
}

func (r *fieldPolicy) check() error {
	return r.policy.Execute(r.field)
}
