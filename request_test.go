package protopolicy

import (
	"context"
	"testing"

	protopolicyv1 "buf.build/gen/go/signal426/protopolicy/protocolbuffers/go/protopolicy/v1"
	"github.com/signal426/protopolicy/policy"
	"github.com/stretchr/testify/assert"
)

func firstNameNotBob(v any) bool {
	return v == "bob"
}

func TestFieldPolicies(t *testing.T) {
	t.Run("it should validate non-zero", func(t *testing.T) {
		// arrange
		req := &protopolicyv1.CreateUserRequest{
			User: &protopolicyv1.User{
				FirstName: "Bob",
			},
		}
		p := ForRequest("createUser", req).
			WithFieldPolicy("user.id", policy.NeverZero()).
			WithFieldPolicy("user.first_name", policy.CalculatedTrait(firstNameNotBob))
		// act
		err := p.GetViolations(context.Background())
		// assert
		assert.Error(t, err)
	})
}
