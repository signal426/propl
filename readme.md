## Propl

A way to set policies on proto fields without writing a ton of `if` statements. This was made for readability and reduction of human error -- although
not slow, it achieves a nice API through reflection and recursion, so not the same scale as the `if` statement route.

###  Usage
Example:
```go
// construct with the rpc name, the request message we're validating, and the mask paths (if any)
p := ForRequest("createUser", req, req.GetUpdateMask().GetPaths()...).
	WithFieldPolicy("user.id", NeverZero()). // NeverZero asserts field is not zero in any situation (message or in mask)
	WithFieldPolicy("user.first_name", NeverZeroWhen(InMask). // NeverZeroWhen only executes the check when the condition is met
		And(Calculated(TraitCalculation{ItShouldNot: "be equal to bob", Calculation: firstNameNotBob}))). // custom allows you to pass a custom function (must unpack from the empty interface)
	WithFieldPolicy("user.last_name", NeverZeroWhen(InMask)).
	WithFieldPolicy("user.primary_address", NeverZeroWhen(InMask)).
	WithFieldPolicy("user.primary_address.line1", NeverZeroWhen(InMask))
```
Any field on the message not specified in the request policy does not get evaluated.
