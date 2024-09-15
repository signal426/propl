package propl

import (
	"context"
	"errors"
	"fmt"
	"testing"

	proplv1 "buf.build/gen/go/signal426/propl/protocolbuffers/go/propl/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type MyErrResultHandler struct{}

func (my MyErrResultHandler) SpaghettiOs(uhohs []UhOh) error {
	var errString string
	for _, uhoh := range uhohs {
		errString += fmt.Sprintf("%s: %s\n", uhoh.Field, uhoh.Err)
	}
	return errors.New(errString)
}

func TestFieldPolicies(t *testing.T) {
	t.Run("it should validate non-zero", func(t *testing.T) {
		// arrange
		req := &proplv1.CreateUserRequest{
			User: &proplv1.User{
				FirstName: "Bob",
			},
		}
		p := ForSubject(req, WithCtx(context.Background())).
			HasNonZeroFields("user", "user.id")
		// act
		err := p.E()
		// assert
		assert.Error(t, err)
		// assert.Equal(t, "")
	})

	t.Run("it should validate not eq", func(t *testing.T) {
		// arrange
		req := &proplv1.CreateUserRequest{
			User: &proplv1.User{
				FirstName: "Bob",
			},
		}
		p := ForSubject(req, WithCtx(context.Background())).
			HasCustomEvaluation("user.first_name", func(ctx context.Context, msg proto.Message) error {
				req, err := AssertType[*proplv1.CreateUserRequest](msg)
				if err != nil {
					return err
				}
				if req.GetUser().GetFirstName() == "Bob" {
					return errors.New("something happened")
				}
				return nil
			})
		// act
		err := p.E()
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
		p := ForSubject(req, WithMaskPaths(req.GetUpdateMask().GetPaths()...)).
			HasNonZeroFields("user.id", "some.fake").
			HasNonZeroFieldsWhen(
				IsInMask("user.last_name"),
				IsInMask("user.primary_address"),
				IsInMask("user.primary_address.line1"))
		// act
		err := p.E()
		// assert
		assert.Error(t, err)
	})

	t.Run("it should validate with custom field infractions handler", func(t *testing.T) {
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
				Paths: []string{"first_name", "last_name", "line1"},
			},
		}
		// ForSubject(request, options...) instantiates the evaluator
		p := ForSubject(req, WithCtx(context.Background()), WithMaskPaths(req.GetUpdateMask().GetPaths()...)).
			// Specify all of the field paths that should not be equal to their zero value
			HasNonZeroFields("user.id", "some.fake").
			// Specify all of the field paths that must not be zero if they meet the specified conditions (e.g. a field is supplied in a mask)
			HasNonZeroFieldsWhen(
				IsInMask("user.first_name"),
				IsInMask("user.last_name"),
				IsInMask("user.primary_address")).
			// specify any custom evaluations that are triggered by the specified field being in chosen conditions
			HasCustomEvaluationWhen(IsInMask("user.primary_address.line1"), func(ctx context.Context, msg proto.Message) error {
				// you don't need to do this -- use type assertions. I do this for the error message
				req, err := AssertType[*proplv1.UpdateUserRequest](msg)
				if err != nil {
					return err
				}
				if req.GetUser().GetPrimaryAddress().GetLine1() == "a" {
					return errors.New("cannot be a")
				}
				return nil
			})
		// act
		// call this before running the evaluation in order to substitute your own error result handler
		// to do things like custom formatting
		err := p.E()
		fmt.Printf("%s\n", err.Error())
		// assert
		assert.Error(t, err)
	})

	t.Run("it should validate with precheck", func(t *testing.T) {
		// arrange
		authorizeUpdate := func(_ context.Context, msg proto.Message) error {
			req, err := AssertType[*proplv1.UpdateUserRequest](msg)
			if err != nil {
				return err
			}
			if req.GetUser().GetId() != "abc123" {
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
		p := ForSubject(req,
			WithCtx(context.Background()),
			WithMaskPaths(req.GetUpdateMask().GetPaths()...),
			WithPrePolicyEvaluation(authorizeUpdate)).
			HasNonZeroFields("user.id", "some.fake").
			HasNonZeroFieldsWhen(
				IsInMask("user.first_name"),
				IsInMask("user.last_name"),
				IsInMask("user.primary_address"))
		// act
		err := p.E()
		fmt.Printf("%s\n", err.Error())
		// assert
		assert.Error(t, err)
	})

	t.Run("it should validate with postcheck", func(t *testing.T) {
		// arrange
		authorizeUpdate := func(_ context.Context, msg proto.Message) error {
			req, err := AssertType[*proplv1.UpdateUserRequest](msg)
			if err != nil {
				return err
			}
			if req.GetUser().GetId() != "abc123" {
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
		p := ForSubject(req,
			WithCtx(context.Background()),
			WithMaskPaths(req.GetUpdateMask().GetPaths()...),
			WithPostPolicyEvaluation(authorizeUpdate)).
			HasNonZeroFields("user.id", "some.fake").
			HasNonZeroFieldsWhen(
				IsInMask("user.first_name"),
				IsInMask("user.last_name"),
				IsInMask("user.primary_address"))
		// act
		err := p.E()
		fmt.Printf("%s\n", err.Error())
		// assert
		assert.Error(t, err)
	})

	t.Run("it should keep processing if global continue is true", func(t *testing.T) {
		// arrange
		authorizeUpdate := func(_ context.Context, msg proto.Message) error {
			req, err := AssertType[*proplv1.UpdateUserRequest](msg)
			if err != nil {
				return err
			}
			if req.GetUser().GetId() != "abc123" {
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
		p := ForSubject(req,
			WithCtx(context.Background()),
			WithContinueOnGlobalEvalErr(),
			WithMaskPaths(req.GetUpdateMask().GetPaths()...),
			WithPostPolicyEvaluation(authorizeUpdate)).
			HasNonZeroFields("user.id", "some.fake").
			HasNonZeroFieldsWhen(
				IsInMask("user.first_name"),
				IsInMask("user.last_name"),
				IsInMask("user.primary_address"))
		// act
		err := p.E()
		fmt.Printf("%s\n", err.Error())
		// assert
		assert.Error(t, err)
	})
}
