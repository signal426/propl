package propl

import (
	"errors"
)

var ErrTypeAssertionFailed = errors.New("type assertion failed")

// No one really needs to use this. Go type assertions are fine.
// This just wraps that in a function if you like using things
// that way.
func AssertType[T interface{}](i interface{}) (T, error) {
	var t T
	t, ok := i.(T)
	if !ok {
		return t, ErrTypeAssertionFailed
	}
	return t, nil
}
