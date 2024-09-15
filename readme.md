## Propl

A way to set policies on proto fields without writing a ton of `if` statements. This was made for readability and reduction of human error -- although
not slow, it achieves a nice API through reflection and recursion, so not the same scale as the `if` statement route.

###  Usage
Example:
```go
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
// call this before running the evaluation in order to substitute your own error result handler
// to do things like custom formatting
err := p.CustomErrResultHandler(MyErrResultHandler{}).E()
```
Any field on the message not specified in the request policy does not get evaluated.
