package hclread

import (
	"context"
	"strings"

	"github.com/go-faster/errors"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/zclconf/go-cty/cty"
)

// load json or yaml schema file

type ValidationError struct {
	*jsonschema.ValidationError
	Location string
	Problems []string
	Range    *hcl.Range
}

func LoadValidationErrors(ctx context.Context, cnt hcl.Expression, errv error) (*ValidationError, hcl.Diagnostics) {

	if verr, ok := errors.Into[*jsonschema.ValidationError](errv); ok {
		ValidationErr := &ValidationError{
			ValidationError: verr,
		}
		for _, cause := range verr.BasicOutput().Errors {
			ValidationErr.Location = cause.InstanceLocation
			ValidationErr.Problems = append(ValidationErr.Problems, cause.Error)
		}

		rng, err := InstanceLocationStringToHCLRange(ValidationErr.Location, cnt)
		if err != nil {
			return nil, err
		}
		ValidationErr.Range = rng
		return ValidationErr, nil
	}

	return nil, hcl.Diagnostics{
		&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid expression",
			Detail:   "unable to find instance location",
		},
	}
}

func InstanceLocationStringToHCLRange(instLoc string, cnt hcl.Expression) (*hcl.Range, hcl.Diagnostics) {
	splt := strings.Split(strings.TrimPrefix(instLoc, "/"), "/")
	return InstanceLocationToHCLRange(splt, cnt)
}

func InstanceLocationToHCLRange(splt []string, cnt hcl.Expression) (*hcl.Range, hcl.Diagnostics) {

	if a, err := cnt.(*hclsyntax.ObjectConsExpr); err {
		for _, item := range a.Items {
			ectx := &hcl.EvalContext{}
			v, err := item.KeyExpr.Value(ectx)
			if err != nil {
				return nil, err
			}

			if v.Type() == cty.String {
				if v.AsString() == splt[0] {
					if len(splt) == 1 {
						r := item.KeyExpr.Range()
						return &r, nil
					}
					return InstanceLocationToHCLRange(splt[1:], item.ValueExpr)
				}
			}
		}

	}

	return nil, hcl.Diagnostics{
		&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid expression",
			Detail:   "unable to find instance location",
			Subject:  cnt.Range().Ptr(),
		},
	}

}