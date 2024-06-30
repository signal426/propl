## Propl

A way to set policies on proto fields without writing a ton of `if` statements. This was made for readability and reduction of human error -- although
not slow, it achieves a nice API through reflection and recursion, so not the same scale as the `if` statement route.

###  Usage
Example:
```go
// construct with the rpc name, the request message we're validating, and the mask paths (if any)
err := propl.ForRequest("createUser", req, req.GetUpdateMask().GetPaths()...).
	WithFieldPolicy("user.id", propl.NeverZero()). // NeverZero asserts field is not zero in any situation (message or in mask)
	WithFieldPolicy("user.first_name", propl.NeverZeroWhen(propl.InMask)). // NeverZeroWhen only executes the check when the condition is met
		And(propl.Calculated("it should not be equal to bob", firstNameNotBob))). // custom allows you to pass a custom function (must unpack from the empty interface)
	WithFieldPolicy("user.last_name", propl.NeverZeroWhen(propl.InMask)).
	WithFieldPolicy("user.primary_address", propl.NeverZeroWhen(propl.InMask)).
	WithFieldPolicy("user.primary_address.line1", propl.NeverZeroWhen(propl.InMask)).E(ctx) // E shorthand for Evaluate()
```
Any field on the message not specified in the request policy does not get evaluated.
