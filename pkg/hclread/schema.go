package hclread

import (
	"context"
	"strconv"
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

func LoadValidationErrors(ctx context.Context, cnt hcl.Expression, ectx *hcl.EvalContext, errv error) ([]*ValidationError, bool) {

	berr := errv
	for errors.Unwrap(berr) != nil {
		berr = errors.Unwrap(berr)
	}

	vers := make([]*ValidationError, 0)

	if verr, ok := errors.Into[*jsonschema.ValidationError](berr); ok {

		for _, cause := range verr.Causes {
			if ve, _ := LoadValidationErrors(ctx, cnt, ectx, cause); ve != nil {
				// basically, if one of our children has an error,
				vers = append(vers, ve...)
			}
		}

		rng, err := InstanceLocationStringToHCLRange(verr.InstanceLocation, cnt, ectx)
		if err != nil {
			return vers, true
		}

		validationErr := &ValidationError{
			ValidationError: verr,
			Range:           rng,
		}

		return append(vers, validationErr), true
	}

	return nil, false

}

func InstanceLocationStringToHCLRange(instLoc string, cnt hcl.Expression, ectx *hcl.EvalContext) (*hcl.Range, hcl.Diagnostics) {
	splt := strings.Split(strings.TrimPrefix(instLoc, "/"), "/")
	return InstanceLocationToHCLRange(splt, cnt, ectx)
}

func InstanceLocationToHCLRange(splt []string, cnt hcl.Expression, ectx *hcl.EvalContext) (*hcl.Range, hcl.Diagnostics) {

	switch t := cnt.(type) {
	case *hclsyntax.ObjectConsExpr:
		{
			for _, item := range t.Items {
				v, err := item.KeyExpr.Value(ectx)
				if err != nil {
					return nil, err
				}

				if v.Type() == cty.String {
					if v.AsString() == splt[0] {
						if len(splt) == 1 {
							r := item.ValueExpr.Range()
							return &r, nil
						}
						return InstanceLocationToHCLRange(splt[1:], item.ValueExpr, ectx)
					}
				}
			}
		}
	case *hclsyntax.TupleConsExpr:
		{
			intr, err := strconv.Atoi(splt[0])
			if err != nil {
				return nil, hcl.Diagnostics{
					&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Invalid expression",
						Detail:   "unexpected tuple index",
						Subject:  cnt.Range().Ptr(),
					},
				}
			}
			return InstanceLocationToHCLRange(splt[1:], t.Exprs[intr], ectx)
		}
	case *hclsyntax.ScopeTraversalExpr:
		{
			// debug when a value hits here, it means it is a block reference
			// will need to traverse the ectx to find the variable this is referencing

			// OR just return the range of this block as the error,
			// but then we need to seperatly validate the other block (in the test the step.checkout block)
			return &t.SrcRange, nil

			// tbh we prob need to return an array in this func, and its parent. think multiple validation errors
		}
	case *hclsyntax.IndexExpr:
		{
		}
	case *hclsyntax.TemplateExpr:
		{
		}
	}

	return nil, hcl.Diagnostics{
		&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid expression",
			Detail:   "unable to find instance loc",
			Subject:  cnt.Range().Ptr(),
		},
	}

}
