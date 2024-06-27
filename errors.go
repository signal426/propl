package protopolicy

import (
	"fmt"
)

type (
	ErrNoValidationFnProvided  error
	ErrMustNotBeEmpty          error
	ErrMustNotBeEmptyAndInMask error
	ErrMustNotBeEmptyIfInMask  error
	ErrMustNotEqual            error
	ErrMustNotEqualIfInMask    error
)

func (r *ProtoPolicy[T, U]) noValidationFnProvided(property string) ErrNoValidationFnProvided {
	return fmt.Errorf("%s is configured for custom validation, but no validation func was provided", property)
}

func mustNotBeEmpty(property string) ErrMustNotBeEmpty {
	return fmt.Errorf("%s must not be empty", property)
}

func mustNotBeEmptyAndInMask(property string) ErrMustNotBeEmptyAndInMask {
	return fmt.Errorf("%s must not be empty if supplied in field mask", property)
}

func mustNotBeZeroIfInMask(property string) ErrMustNotBeEmptyIfInMask {
	return fmt.Errorf("%s must not be empty and supplied in field mask", property)
}

func mustNotEqual(property, expected, actual string) ErrMustNotEqual {
	return fmt.Errorf("unexpected type for %s, got %s wanted %s", property, actual, expected)
}
