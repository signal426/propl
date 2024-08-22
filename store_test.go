package propl

import (
	"testing"

	proplv1 "buf.build/gen/go/signal426/propl/protocolbuffers/go/propl/v1"
	"github.com/stretchr/testify/assert"
)

func TestCreateFieldStoreFromMessage(t *testing.T) {
	t.Run("it should hydrate a field store", func(t *testing.T) {
		// arrange
		msg := &proplv1.CreateUserRequest{
			User: &proplv1.User{
				FirstName: "bob",
				LastName:  "loblaw",
			},
		}
		// act
		s := newFieldStore[*proplv1.CreateUserRequest](msg)
		s.processPath("user.first_name").
			processPath("user.id").
			processPath("user.last_name")
		// assert
		id := s.dataAtPath("user.id")
		fn := s.dataAtPath("user.first_name")
		ln := s.dataAtPath("user.last_name")
		assert.False(t, id.s(), "id should not be set")
		assert.True(t, id.z(), "id should be zero")
		assert.True(t, fn.s(), "first name should be set")
		assert.Equal(t, fn.v(), "bob", "first name should be bob")
		assert.True(t, ln.s(), "last name should be set")
		assert.Equal(t, ln.v(), "loblaw", "last name should be equal")
	})

	t.Run("it should hydrate a complex field store", func(t *testing.T) {
		// arrange
		msg := &proplv1.CreateUserRequest{
			User: &proplv1.User{
				Id:        "abc123",
				FirstName: "bob",
				LastName:  "loblaw",
				PrimaryAddress: &proplv1.Address{
					Line1: "321",
					Line2: "dddd",
				},
				SecondaryAddresses: []*proplv1.Address{
					{
						Line1: "rrrr",
						Line2: "fvvvv",
					},
				},
			},
		}
		// act
		s := newFieldStore[*proplv1.CreateUserRequest](msg)
		s.processPath("user.id").
			processPath("user.first_name").
			processPath("user.last_name").
			processPath("user.primary_address").
			processPath("user.primary_address.line1").
			processPath("user.primary_address.line2")
		// assert
		id := s.dataAtPath("user.id")
		fn := s.dataAtPath("user.first_name")
		ln := s.dataAtPath("user.last_name")
		pa := s.dataAtPath("user.primary_address")
		pal1 := s.dataAtPath("user.primary_address.line1")
		pal2 := s.dataAtPath("user.primary_address.line2")
		assert.True(t, id.s(), "id should be set")
		assert.Equal(t, id.v(), "abc123", "id should be abc123")
		assert.True(t, fn.s(), "first name should be set")
		assert.Equal(t, fn.v(), "bob", "first name should be bob")
		assert.True(t, ln.s(), "last name should be set")
		assert.Equal(t, ln.v(), "loblaw", "last name should be equal")
		assert.NotNil(t, pa.v(), "primary address should not be nil")
		assert.Equal(t, pal1.v(), "321", "primary address line 1 should be 321")
		assert.Equal(t, pal2.v(), "dddd", "primary address line 2 should be dddd")
	})
}
