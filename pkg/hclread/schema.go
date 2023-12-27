package hclread

import (
	"context"
	"fmt"
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
	// *jsonschema.ValidationError
	// Location string
	Problems []string
	Range    *hcl.Range
}

func DiagnosticToValidationError(ctx context.Context, diags hcl.Diagnostics) ([]*ValidationError, error) {

	vers := make([]*ValidationError, 0)

	for _, diag := range diags {
		vers = append(vers, &ValidationError{
			// Location: diag.Subject,
			Problems: []string{diag.Detail},
			Range:    diag.Subject,
		})
	}
	return vers, nil

}

func LoadValidationErrors(ctx context.Context, cnt hclsyntax.Expression, ectx *hcl.EvalContext, errv error, bdy hcl.Body) (hcl.Diagnostics, error) {

	berr := errv
	for errors.Unwrap(berr) != nil {
		berr = errors.Unwrap(berr)
	}

	fmt.Println("HERE", berr)

	diags := hcl.Diagnostics{}

	if verr, ok := terrors.Into[*jsonschema.ValidationError](berr); ok {

		if len(verr.Causes) > 0 {
			for _, cause := range verr.Causes {
				if ve, err := LoadValidationErrors(ctx, cnt, ectx, cause, bdy); err != nil {
					return nil, err
				} else {
					// basically, if one of our children has an error,
					for _, v := range ve {
						diags = append(diags, v)
					}
				}
			}
		} else {
			rng, diagd := InstanceLocationStringToHCLRange(verr.InstanceLocation, verr.Message, cnt, ectx, bdy)
			if diagd.HasErrors() {
				return nil, diagd
			}

			diag := &hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     verr.Message,
				Detail:      verr.Message,
				Subject:     rng.Range().Ptr(),
				Expression:  rng,
				EvalContext: ectx,
			}

			diags = append(diags, diag)
		}

		// return append(vers, validationErr), nil
	}

	return diags, nil

}

func InstanceLocationStringToHCLRange(instLoc string, msg string, cnt hclsyntax.Expression, ectx *hcl.EvalContext, file hcl.Body) (hcl.Expression, hcl.Diagnostics) {
	splt := strings.Split(strings.TrimPrefix(instLoc, "/"), "/")

	cmp := regexp.MustCompile("additionalProperties '(.*)' not allowed")
	matches := cmp.FindStringSubmatch(msg)
	if len(matches) == 2 {
		splt = append(splt, matches[1])
	}

	return roll2(splt, cnt, ectx, file)
}

func roll2(splt []string, e hcl.Expression, ectx *hcl.EvalContext, file hcl.Body) (hcl.Expression, hcl.Diagnostics) {
	// pp.Println(splt)
	if len(splt) == 0 {
		return e, nil
	}

	if x, ok := e.(*hclsyntax.ObjectConsExpr); ok {
		for _, rr := range x.Items {
			kvf, diags := rr.KeyExpr.Value(ectx)
			if diags.HasErrors() {
				return nil, diags
			}

			if kvf.AsString() == splt[0] {
				if len(splt) == 1 {
					return rr.KeyExpr, nil
				}
				return roll2(splt[1:], rr.ValueExpr, ectx, file)
			}
		}
		return e, nil
	} else if x, ok := e.(*hclsyntax.TupleConsExpr); ok {

		intr, err := strconv.Atoi(splt[0])
		if err != nil {
			for _, exp := range x.Exprs {
				ex, diags := roll2(splt, exp, ectx, file)
				if diags.HasErrors() {
					// todo maybe not
					return nil, diags
				}
				if err == nil {
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

		// hclsyntax.VisitAll(x, func(node hclsyntax.Node) hcl.Diagnostics {
		// 	pp.Println("HERE", node, node.Range())
		// 	return nil
		// })

		// // pp.Println(x.Traversal)
		// ttrav, diag := hcl.AbsTraversalForExpr(x)
		// if diag.HasErrors() {
		// 	return nil, hcl.Diagnostics{
		// 		&hcl.Diagnostic{
		// 			Severity: hcl.DiagError,
		// 			Summary:  "Invalid expression",
		// 			Detail:   "unable to find instance loc",
		// 			Subject:  x.Range().Ptr(),
		// 		},
		// 	}
		// }

		name := x.Traversal.RootName()

		labs := []string{}

		for _, lab := range x.Traversal {
			if z, ok := lab.(hcl.TraverseAttr); ok {
				labs = append(labs, z.Name)
			}
		}

		// pp.Println("foundg", labs, splt)

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
					splt = append(labs[len(blk.Labels):], splt...)
					// pp.Println("found block", blk.Type, blk.Labels, labs, splt)
					for zz, k := range blk.Body.Attributes {
						if zz == splt[0] {
							// pp.Println("hello", k.Expr.Range(), k.Range())
							if len(splt) == 1 {
								// pp.Println(k.Expr.Range())
								// pp.Println(k.Range())
								return k.Expr, nil
							}
							return roll2(splt[1:], k.Expr, ectx, blk.Body)
						}
					}
				}
			}
		}

	}

	return e, nil

}

// func (me *FileBlockEvaluation) GetJSONSchema(ctx context.Context) (*jsonschema.Schema, error) {
// 	s, err := schemas.LoadJSONSchema(ctx, me.Schema)
// 	if err != nil {
// 		return s, terrors.Wrap(err, "problem getting schema").Event(func(e *zerolog.Event) *zerolog.Event {
// 			return e.Int("schema_size", len(me.Schema))
// 		})
// 	}

// 	return s, nil
// }

// func (me *FileBlockEvaluation) GetProperties(ctx context.Context) (map[string]*jsonschema.Schema, error) {
// 	schema, err := me.GetJSONSchema(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	m := make(map[string]*jsonschema.Schema)

// 	getAllDefs("root", schema, m)

// 	return m, nil
// }

// func (me *FileBlockEvaluation) ValidateJSONSchemaProperty(ctx context.Context, prop string) error {
// 	if prop == MetaKey {
// 		return nil
// 	}

// 	schema, err := me.GetProperties(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	if schema[prop] == nil {
// 		return terrors.Errorf("property %q not found", prop)
// 	}

// 	return schema[prop].Validate(me.RawOutput)

// }

func getAllDefs(parent string, schema *jsonschema.Schema, defs map[string]*jsonschema.Schema) {
	////////////////////////////////////////
	// not sure if this is needed or not
	if defs[parent] != nil {
		return
	}
	////////////////////////////////////////

	if schema == nil {
		return
	}

	for k, v := range schema.DependentSchemas {
		if ok := defs[k]; ok != nil {
			continue
		}
		defs[k] = v
		getAllDefs(k, v, defs)
	}

	for k, v := range schema.Properties {
		if ok := defs[k]; ok != nil {
			continue
		}
		defs[k] = v
		getAllDefs(k, v, defs)
	}

	for _, v := range schema.AllOf {
		getAllDefs(parent, v, defs)
	}

	for _, v := range schema.AnyOf {
		getAllDefs(parent, v, defs)
	}

	for _, v := range schema.OneOf {
		getAllDefs(parent, v, defs)
	}

	for _, v := range schema.PatternProperties {
		getAllDefs(parent, v, defs)
	}

	switch v := schema.Items.(type) {
	case *jsonschema.Schema:
		getAllDefs(parent, v, defs)
	case []*jsonschema.Schema:
		for _, v := range v {
			getAllDefs(parent, v, defs)
		}
	}
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
