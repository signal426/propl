package propl

import (
	"testing"

	proplv1 "buf.build/gen/go/signal426/propl/protocolbuffers/go/propl/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestCreateFieldStoreFromMessage(t *testing.T) {
	t.Run("it should create a field store from a proto message", func(t *testing.T) {
		// arrange
		expected := fieldStore{
			"user.id": {
				zero: false,
				path: "user.id",
				val:  "abc123",
				set:  true,
			},
			"user.first_name": {
				zero:   true,
				path:   "first_name",
				val:    nil,
				inMask: true,
			},
			"update_mask": {
				zero: false,
				path: "update_mask",
				val: &fieldmaskpb.FieldMask{
					Paths: []string{"first_name"},
				},
				set: true,
			},
			"update_mask.paths": {
				zero: false,
				path: "update_mask.paths",
				val:  []string{"first_name"},
				set:  true,
			},
		}
		input := &proplv1.UpdateUserRequest{
			User: &proplv1.User{
				Id: "abc123",
			},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"first_name"},
			},
		}
		// act
		store := newFieldStore()
		store.fill(input, "first_name")
		// assert
		assert.Equal(t, 5, len(store))
		for _, e := range expected {
			a := store.getByPath(e.p())
			assert.Equal(t, e.p(), a.p())
			assert.Equal(t, e.z(), a.z())
			assert.Equal(t, e.s(), a.s())
			assert.Equal(t, e.m(), a.m())
		}
	})
}
