package hclread

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-faster/errors"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/walteh/terrors"
)

// load json or yaml schema file

type ValidationError struct {
	*jsonschema.ValidationError
	Location string
	Problems []string
	Range    *hcl.Range
}

func LoadValidationErrors(ctx context.Context, cnt hclsyntax.Expression, ectx *hcl.EvalContext, errv error, bdy hcl.Body) ([]*ValidationError, error) {

	berr := errv
	for errors.Unwrap(berr) != nil {
		berr = errors.Unwrap(berr)
	}

	vers := make([]*ValidationError, 0)

	if verr, ok := terrors.Into[*jsonschema.ValidationError](berr); ok {

		for _, cause := range verr.Causes {
			if ve, err := LoadValidationErrors(ctx, cnt, ectx, cause, bdy); err != nil {
				return nil, err
			} else {
				// basically, if one of our children has an error,
				vers = append(vers, ve...)
			}
		}

		rng, err := InstanceLocationStringToHCLRange(verr.InstanceLocation, verr.Message, cnt, ectx, bdy)
		if err != nil {
			return nil, err
		}

		validationErr := &ValidationError{
			ValidationError: verr,
			Range:           rng,
		}

		return append(vers, validationErr), nil
	}

	return vers, nil

}

func InstanceLocationStringToHCLRange(instLoc string, msg string, cnt hclsyntax.Expression, ectx *hcl.EvalContext, file hcl.Body) (*hcl.Range, error) {
	splt := strings.Split(strings.TrimPrefix(instLoc, "/"), "/")

	cmp := regexp.MustCompile("additionalProperties '(.*)' not allowed")
	matches := cmp.FindStringSubmatch(msg)
	if len(matches) == 2 {
		splt = append(splt, matches[1])
	}

	return roll2(splt, cnt, ectx, file)
}

func roll2(splt []string, e hcl.Expression, ectx *hcl.EvalContext, file hcl.Body) (*hcl.Range, error) {
	if len(splt) == 0 {
		return e.Range().Ptr(), nil
	}

	if x, ok := e.(*hclsyntax.ObjectConsExpr); ok {
		for _, rr := range x.Items {
			kvf, err := rr.KeyExpr.Value(ectx)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to evaluate %q", rr.KeyExpr)
			}
			// rz, errd := roll2(splt, rr.ValueExpr, ectx, file)
			// if errd != nil {
			// 	return nil, errd
			// }
			// fmt.Println(" ----- ", kvf.AsString(), splt, rz)
			// if rz == nil {
			// 	continue
			// }

			if kvf.AsString() == splt[0] {
				if len(splt) == 1 {
					return rr.KeyExpr.Range().Ptr(), nil
				}
				return roll2(splt[1:], rr.ValueExpr, ectx, file)
			}
		}
		return e.Range().Ptr(), nil
	} else if x, ok := e.(*hclsyntax.TupleConsExpr); ok {
		// wrk := make([]any, 0)
		// for _, exp := range x.Exprs {
		// 	r, err := roll2(splt, exp, ectx, file)
		// 	if err != nil {
		// 		return nil, terrors.Wrapf(err, "failed to evaluate %q", exp)
		// 	}
		// 	if r == nil {
		// 		continue
		// 	}
		// 	// wrk = append(wrk, r)
		// }

		// pp.Println(" ----- ", splt, x.Exprs)
		intr, err := strconv.Atoi(splt[0])
		if err != nil {
			for _, exp := range x.Exprs {
				ex, err := roll2(splt, exp, ectx, file)
				if err == nil {
					// return nil, terrors.Wrapf(err, "failed to evaluate %q", exp)
					return ex, nil
				}

				// if err != nil {
				// 	return nil, terrors.Wrapf(err, "failed to evaluate %q", exp)
				// }
			}
		}
		return roll2(splt[1:], x.Exprs[intr], ectx, file)
		// return wrk, nil
	} else if x, ok := e.(*hclsyntax.ScopeTraversalExpr); ok {

		name := x.Traversal.RootName()

		labs := []string{}

		for _, lab := range x.Traversal {
			if z, ok := lab.(hcl.TraverseAttr); ok {
				labs = append(labs, z.Name)
			}
		}

		if bdy, ok := file.(*hclsyntax.Body); ok {
		HERE:
			for _, blk := range bdy.Blocks {
				if blk.Type == name {
					if len(labs) < len(blk.Labels) {
						break HERE
					}
					for i, v := range blk.Labels {
						if v != labs[i] {
							break HERE
						}
					}
					for zz, k := range blk.Body.Attributes {
						if zz == splt[0] {
							if len(splt) == 1 {
								return k.Range().Ptr(), nil
							}
							return roll2(splt[1:], k.Expr, ectx, blk.Body)
						}
					}
				}
			}
		}

		// attrs, err := file.JustAttributes()
		// if err.HasErrors() {
		// 	fmt.Println(attrs)
		// 	return nil, terrors.Wrapf(err, "failed to get attributes")
		// }

		// var attr *hcl.Attribute

		// for _, attrd := range attrs {
		// 	if attrd.Name == name {
		// 		attr = attrd
		// 		break
		// 	}
		// }

		// if attr == nil {
		// 	return nil, errors.Errorf("failed to find block %q", name)
		// }

		// return roll2(splt, attr.Expr, ectx, file)

	}

	// evaled, errd := e.Value(ectx)
	// if errd != nil {
	// 	return nil, errors.Wrapf(errd, "failed to evaluate %q", e)
	// }

	return e.Range().Ptr(), nil

}

// func InstanceLocationToHCLRange(splt []string, cnt hcl.Expression, ectx *hcl.EvalContext) (*hcl.Range, error) {
// 	if len(splt) == 0 {
// 		return cnt.Range().Ptr(), nil
// 	}

// 	// fmt.Println(reflect.TypeOf(cnt))

// 	switch t := cnt.(type) {
// 	case *hclsyntax.ObjectConsExpr:
// 		{
// 			for _, item := range t.Items {
// 				v, err := item.KeyExpr.Value(ectx)
// 				if err != nil {
// 					return nil, err
// 				}

// 				// if v.Type() == cty.String {
// 				// fmt.Println(v.AsString(), splt[0])
// 				// fmt.Println(v.Type(), v.AsString(), splt)
// 				if v.AsString() == splt[0] {
// 					if len(splt) == 1 {
// 						return item.ValueExpr.Range().Ptr(), nil
// 					}
// 					return InstanceLocationToHCLRange(splt[1:], item.ValueExpr, ectx)
// 				}
// 				// }
// 			}
// 		}
// 	case *hclsyntax.TupleConsExpr:
// 		{
// 			intr, err := strconv.Atoi(splt[0])
// 			if err != nil {
// 				return nil, hcl.Diagnostics{
// 					&hcl.Diagnostic{
// 						Severity: hcl.DiagError,
// 						Summary:  "Invalid expression",
// 						Detail:   "unexpected tuple index",
// 						Subject:  cnt.Range().Ptr(),
// 					},
// 				}
// 			}
// 			return InstanceLocationToHCLRange(splt[1:], t.Exprs[intr], ectx)
// 		}
// 	case *hclsyntax.ScopeTraversalExpr:
// 		{
// 			// debug when a value hits here, it means it is a block reference
// 			// will need to traverse the ectx to find the variable this is referencing

// 			// OR just return the range of this block as the error,
// 			// but then we need to seperatly validate the other block (in the test the step.checkout block)
// 			return &t.SrcRange, nil

// 			// tbh we prob need to return an array in this func, and its parent. think multiple validation errors
// 		}
// 	case *hclsyntax.IndexExpr:
// 		{
// 			// TODO: this is just a placeholder, not the right way to do this
// 			return &t.SrcRange, nil
// 		}
// 	case *hclsyntax.TemplateExpr:
// 		{
// 			// TODO: this is just a placeholder, not the right way to do this
// 			return &t.SrcRange, nil
// 		}
// 	}

// 	// return nil, hcl.Diagnostics{
// 	// 	&hcl.Diagnostic{
// 	// 		Severity: hcl.DiagError,
// 	// 		Summary:  "Invalid expression",
// 	// 		Detail:   "unable to find instance loc",
// 	// 		Subject:  cnt.Range().Ptr(),
// 	// 	},
// 	// }

// 	return cnt.Range().Ptr(), nil

// }
