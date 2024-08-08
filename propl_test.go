package propl

import (
	"context"
	"errors"
	"fmt"
	"testing"

	proplv1 "buf.build/gen/go/signal426/propl/protocolbuffers/go/propl/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestFieldPolicies(t *testing.T) {
	t.Run("it should validate non-zero", func(t *testing.T) {
		// arrange
		req := &proplv1.CreateUserRequest{
			User: &proplv1.User{
				FirstName: "Bob",
			},
		}
		p := For(req).NeverZero("user.id")
		// act
		err := p.E(context.Background())
		// assert
		assert.Error(t, err)
	})

	t.Run("it should validate not eq", func(t *testing.T) {
		// arrange
		req := &proplv1.CreateUserRequest{
			User: &proplv1.User{
				FirstName: "Bob",
			},
		}
		p := For(req).NeverEq("user.first_name", "Bob")
		// act
		err := p.E(context.Background())
		// assert
		assert.Error(t, err)
	})

	t.Run("it should validate complex", func(t *testing.T) {
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
		p := For(req, req.GetUpdateMask().GetPaths()...).
			NeverZero("user.id").
			NeverZero("some.fake").
			NeverZeroWhen("user.first_name", InMask).
			NeverZeroWhen("user.last_name", InMask).
			NeverZeroWhen("user.primary_address", InMask).
			NeverZeroWhen("user.primary_address.line1", InMask)
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
		p := For(req, req.GetUpdateMask().GetPaths()...).
			WithFieldInfractionsHandler(finfractionsHandler).
			NeverZero("user.id").
			NeverZero("some.fake").
			NeverZeroWhen("user.first_name", InMask).
			NeverZeroWhen("user.last_name", InMask).
			NeverZeroWhen("user.primary_address", InMask).
			NeverZeroWhen("user.primary_address.line1", InMask)
		// act
		err := p.E(context.Background())
		// assert
		assert.Error(t, err)
	})

	t.Run("it should validate with precheck", func(t *testing.T) {
		// arrange
		authorizeUpdate := func(_ context.Context, msg *proplv1.UpdateUserRequest) error {
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
		p := For(req, req.GetUpdateMask().GetPaths()...).
			WithPrecheckPolicy(authorizeUpdate).
			NeverZero("user.id").
			NeverZero("some.fake").
			NeverZeroWhen("user.first_name", InMask).
			NeverZeroWhen("user.last_name", InMask).
			NeverZeroWhen("user.primary_address", InMask).
			NeverZeroWhen("user.primary_address.line1", InMask)
		// act
		err := p.E(context.Background())
		// assert
		assert.Error(t, err)
	})
}
