package propl

import (
	"context"
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
		s := initFieldStore(msg)
		id, iderr := AssertType[*fieldData](s.ProcessByID(context.Background(), "user.id"))
		fn, fnerr := AssertType[*fieldData](s.ProcessByID(context.Background(), "user.first_name"))
		ln, lnerr := AssertType[*fieldData](s.ProcessByID(context.Background(), "user.last_name"))
		// assert
		assert.NoError(t, iderr)
		assert.NoError(t, fnerr)
		assert.NoError(t, lnerr)
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
		s := initFieldStore(msg)
		ctx := context.Background()
		id, iderr := AssertType[*fieldData](s.ProcessByID(ctx, "user.id"))
		fn, fnerr := AssertType[*fieldData](s.ProcessByID(ctx, "user.first_name"))
		ln, lnerr := AssertType[*fieldData](s.ProcessByID(ctx, "user.last_name"))
		pa, paerr := AssertType[*fieldData](s.ProcessByID(ctx, "user.primary_address"))
		pal1, pal1err := AssertType[*fieldData](s.ProcessByID(ctx, "user.primary_address.line1"))
		pal2, pal2err := AssertType[*fieldData](s.ProcessByID(ctx, "user.primary_address.line2"))
		assert.NoError(t, iderr)
		assert.NoError(t, fnerr)
		assert.NoError(t, lnerr)
		assert.NoError(t, paerr)
		assert.NoError(t, pal1err)
		assert.NoError(t, pal2err)
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
