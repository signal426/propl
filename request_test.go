package propl

import (
	"context"
	"testing"

	proplv1 "buf.build/gen/go/signal426/propl/protocolbuffers/go/propl/v1"
	"github.com/stretchr/testify/assert"
)

func firstNameNotBob(v any) bool {
	return v == "bob"
}

func TestFieldPolicies(t *testing.T) {
	t.Run("it should validate non-zero", func(t *testing.T) {
		// arrange
		req := &proplv1.CreateUserRequest{
			User: &proplv1.User{
				FirstName: "Bob",
			},
		}
		p := ForRequest("createUser", req).
			WithFieldPolicy("user.id", NeverZero()).
			WithFieldPolicy("user.first_name", Calculated(firstNameNotBob))
		// act
		err := p.GetViolations(context.Background())
		// assert
		assert.Error(t, err)
	})

	t.Run("it should validate complex", func(t *testing.T) {
		// arrange
		req := &proplv1.CreateUserRequest{
			User: &proplv1.User{
				FirstName: "Bob",
			},
		}
		p := ForRequest("createUser", req).
			WithFieldPolicy("user.id", NeverZero()).
			WithFieldPolicy("user.address.line1", NeverZeroWhen(InMask)).
			WithFieldPolicy("user.address", NeverZero()).
			WithFieldPolicy("user.first_name", Calculated(firstNameNotBob))
		// act
		err := p.GetViolations(context.Background())
		// fmt.Printf(err.Error())
		// assert
		assert.Error(t, err)
	})
}
