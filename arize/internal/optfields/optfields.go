// Package optfields holds tiny helpers used by subclients to forward optional
// value-typed request fields into the generated body/params types (which use
// pointers for optionality).
package optfields

// PtrIfSet returns &v when v is not the zero value of T, else nil. Use to
// forward an optional value-typed request field into a generated body that
// uses pointers, omitting it when the caller left it zero.
func PtrIfSet[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}

// PtrWithDefault returns &v when v is not the zero value of T, else &fallback.
// Use when the SDK itself substitutes a default value (rather than relying on
// the server to do so).
func PtrWithDefault[T comparable](v, fallback T) *T {
	var zero T
	if v == zero {
		return &fallback
	}
	return &v
}
