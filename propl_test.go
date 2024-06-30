package propl

import (
	"context"
	"errors"
	"fmt"
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
		p := ForMessage(req).
			WithFieldPolicy("user.id", NeverZero())
		// act
		err := p.E(context.Background())
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
		p := ForMessage(req).
			WithFieldPolicy("user.first_name", Calculated("be equal to bob", firstNameNotBob))
		// act
		err := p.E(context.Background())
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
				Paths: []string{"first_name", "last_name"},
			},
		}
		p := ForMessage(req, req.GetUpdateMask().GetPaths()...).
			WithFieldPolicy("user.id", NeverZero()).
			WithFieldPolicy("some.fake", NeverZero()).
			WithFieldPolicy("user.first_name", NeverZeroWhen(InMask).
				And(Calculated("it should not be equal to bob", firstNameNotBob))).
			WithFieldPolicy("user.last_name", NeverZeroWhen(InMask)).
			WithFieldPolicy("user.primary_address", NeverZeroWhen(InMask)).
			WithFieldPolicy("user.primary_address.line1", NeverZeroWhen(InMask))
		// act
		err := p.E(context.Background())
		// assert
		assert.Error(t, err)
	})

	t.Run("it should validate with custom field infractions handler", func(t *testing.T) {
		// arrange
		finfractionsHandler := func(i map[string]error) error {
			var errString string
			for k, v := range i {
				errString += fmt.Sprintf("%s: %s\n", k, v.Error())
			}
			return errors.New(errString)
		}
		req := &proplv1.UpdateUserRequest{
			User: &proplv1.User{
				FirstName: "bob",
				PrimaryAddress: &proplv1.Address{
					Line1: "a",
					Line2: "b",
				},
			},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"first_name", "last_name"},
			},
		}
		p := ForMessage(req, req.GetUpdateMask().GetPaths()...).
			WithFieldInfractionsHandler(finfractionsHandler).
			WithFieldPolicy("user.id", NeverZero()).
			WithFieldPolicy("some.fake", NeverZero()).
			WithFieldPolicy("user.first_name", NeverZeroWhen(InMask).
				And(Calculated("it should not be equal to bob", firstNameNotBob))).
			WithFieldPolicy("user.last_name", NeverZeroWhen(InMask)).
			WithFieldPolicy("user.primary_address", NeverZeroWhen(InMask)).
			WithFieldPolicy("user.primary_address.line1", NeverZeroWhen(InMask))
		// act
		err := p.E(context.Background())
		// assert
		assert.Error(t, err)
	})

	t.Run("it should validate with precheck", func(t *testing.T) {
		// arrange
		authorizeUpdate := func(msg *proplv1.UpdateUserRequest) error {
			if msg.GetUser().GetId() != "abc123" {
				return errors.New("can only update user abc123")
			}
			return nil
		}
		req := &proplv1.UpdateUserRequest{
			User: &proplv1.User{
				FirstName: "bob",
				PrimaryAddress: &proplv1.Address{
					Line1: "a",
					Line2: "b",
				},
			},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"first_name", "last_name"},
			},
		}
		p := ForMessage(req, req.GetUpdateMask().GetPaths()...).
			WithPrecheckPolicy(authorizeUpdate).
			WithFieldPolicy("user.id", NeverZero()).
			WithFieldPolicy("some.fake", NeverZero()).
			WithFieldPolicy("user.first_name", NeverZeroWhen(InMask).
				And(Calculated("it should not be equal to bob", firstNameNotBob))).
			WithFieldPolicy("user.last_name", NeverZeroWhen(InMask)).
			WithFieldPolicy("user.primary_address", NeverZeroWhen(InMask)).
			WithFieldPolicy("user.primary_address.line1", NeverZeroWhen(InMask))
		// act
		err := p.E(context.Background())
		// assert
		assert.Error(t, err)
	})
}
