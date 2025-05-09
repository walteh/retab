package diff

import "github.com/google/go-cmp/cmp"

// TestingOpts contains options for diff testing functionality
//
//go:opts
type TestingOpts struct {
	cmpOpts          []cmp.Option
	logRawDiffOnFail bool
}

// WithUnexportedType adds an option to allow comparing unexported fields in a type
// This is useful when testing struct values with private fields
func WithUnexportedType[T any]() OptTestingOptsSetter {
	return func(opts *TestingOpts) {
		var v T
		opts.cmpOpts = append(opts.cmpOpts, cmp.AllowUnexported(v))
	}
}
