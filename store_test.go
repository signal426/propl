package propl

import (
	"testing"

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
			},
			"user.first_name": {
				zero: true,
				path: "first_name",
				val:  nil,
			},
			"update_mask": {
				zero: false,
				path: "update_mask",
				val: &fieldmaskpb.FieldMask{
					Paths: []string{"first_name"},
				},
			},
			"update_mask.paths": {
				zero: false,
				path: "update_mask.paths",
				val:  []string{"first_name"},
			},
		}
		input := &protopolicyv1.UpdateUserRequest{
			User: &protopolicyv1.User{
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
		}
	})
}
