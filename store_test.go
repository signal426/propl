package protopolicy

import (
	"fmt"
	"testing"

	protopolicyv1 "buf.build/gen/go/signal426/protopolicy/protocolbuffers/go/protopolicy/v1"
	"github.com/stretchr/testify/assert"
)

func TestCreateFieldStoreFromMessage(t *testing.T) {
	t.Run("it should create a field store from a proto message", func(t *testing.T) {
		// arrange
		expected := fieldStore{
			"user.id": {
				zero: true,
				path: "user.id",
				val:  "",
			},
			"user.last_name": {
				zero: false,
				path: "user.last_name",
				val:  "test",
			},
		}
		input := &protopolicyv1.User{
			Id:       "test",
			LastName: "test",
		}
		// act
		store := messageToFieldStore(input, ".")
		fmt.Printf("%+v\n", store)
		// assert
		assert.Equal(t, 2, len(store))
		for _, e := range expected {
			a := store.getByPath(e.p())
			assert.Equal(t, e.p(), a.p())
			assert.Equal(t, e.v(), a.v())
			assert.Equal(t, e.z(), a.z())
		}
	})
}
