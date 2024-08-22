## Propl

A way to set policies on proto fields without writing a ton of `if` statements. This was made for readability and reduction of human error -- although
not slow, it achieves a nice API through reflection and recursion, so not the same scale as the `if` statement route.

###  Usage
Example:
```go
// can specify a precheck function (e.g. an authz check) that will skip the rest of the evaluations if fails
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
err := p.E(context.Background())
```
Any field on the message not specified in the request policy does not get evaluated.
