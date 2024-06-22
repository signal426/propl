package protovalidate

import (
	"fmt"
)

type (
	ErrNoValidationFnProvided error
	ErrMustNotBeEmpty         error
	ErrMustNotBeEmptyIfInMask error
	ErrMustNotEqual           error
	ErrMustNotEqualIfInMask   error
)

func noValidationFnProvided(property, rpc string) ErrNoValidationFnProvided {
	return fmt.Errorf("%s is configured for custom validation, but no validation func was provided")
}

// todo: make rpc optional
func mustNotBeEmpty(property, rpc string) ErrMustNotBeEmpty {
	return fmt.Errorf("%s must not be empty for %s", property, rpc)
}

func mustNotBeEmptyIfInMask(property, rpc string) ErrMustNotBeEmptyIfInMask {
	return fmt.Errorf("%s must not be empty if supplied in field mask for %s", property, rpc)
}

func mustBeSuppliedInMaskAndNonZero(property, rpc string) ErrMustNotBeEmptyIfInMask {
	return fmt.Errorf("%s must not be empty and supplied in field mask for %s", property, rpc)
}

func mustNotEqual(property, rpc, expected, actual string) ErrMustNotEqual {
	return fmt.Errorf("unexpected type for %s, got %s wanted %s for %s", property, actual, expected, rpc)
}
