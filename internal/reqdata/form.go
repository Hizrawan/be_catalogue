package reqdata

// Form represents a request definition, with rules how to check for authorization and validation of the value.
// This interface can be used with a Controller. If provided to Validate function of a Controller, it will perform
// the checks according to the method implementation and return proper HTTP response in case of failed checks.
type Form interface {
	Authorized(ctx *Context) bool
	Validate(ctx *Context) error
}
