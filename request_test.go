package propl

import (
	"context"
	"strings"
	"testing"

	proplv1 "buf.build/gen/go/signal426/propl/protocolbuffers/go/propl/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func firstNameNotBob(v any) bool {
	return strings.ToLower(v.(string)) != "bob"
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
			WithFieldPolicy("user.id", NeverZero())
		// act
		err := p.GetViolations(context.Background())
		// assert
		assert.Error(t, err)
	})

	t.Run("it should validate custom", func(t *testing.T) {
		// arrange
		req := &proplv1.CreateUserRequest{
			User: &proplv1.User{
				FirstName: "Bob",
			},
		}
		p := ForRequest("createUser", req).
			WithFieldPolicy("user.first_name", Calculated("be equal to bob", firstNameNotBob))
		// act
		err := p.GetViolations(context.Background())
		// assert
		assert.Error(t, err)
	})

	t.Run("it should validate nested", func(t *testing.T) {
		// arrange
		req := &proplv1.UpdateUserRequest{
			User: &proplv1.User{
				FirstName: "bob",
				PrimaryAddress: &proplv1.Address{
					Line1: "a",
					Line2: "b",
				},
			},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"first_name"},
			},
		}
		p := ForRequest("createUser", req, req.GetUpdateMask().GetPaths()...).
			WithFieldPolicy("user.id", NeverZero()).
			WithFieldPolicy("user.first_name", NeverZeroWhen(InMask).
				And(Calculated("it should not be equal to bob", firstNameNotBob))).
			WithFieldPolicy("user.last_name", NeverZeroWhen(InMask)).
			WithFieldPolicy("user.primary_address", NeverZeroWhen(InMask)).
			WithFieldPolicy("user.primary_address.line1", NeverZeroWhen(InMask))
		// act
		err := p.GetViolations(context.Background())
		// assert
		assert.Error(t, err)
	})
}
