package lang

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

func LoadValidationErrors(ctx context.Context, cnt hclsyntax.Expression, ectx *hcl.EvalContext, errv error, bdy *BodyBuilder, sudo *SudoContext) (hcl.Diagnostics, error) {

	berr := errv
	for errors.Unwrap(berr) != nil {
		berr = errors.Unwrap(berr)
	}

	diags := hcl.Diagnostics{}

	if verr, ok := terrors.Into[*jsonschema.ValidationError](berr); ok {

		if len(verr.Causes) > 0 {
			for _, cause := range verr.Causes {
				if ve, err := LoadValidationErrors(ctx, cnt, ectx, cause, bdy, sudo); err != nil {
					return nil, err
				} else {
					// basically, if one of our children has an error,
					diags = append(diags, ve...)
				}
			}
		} else {
			ctxd, diagd := InstanceLocationStringToHCLRange(verr.InstanceLocation, verr.Message, cnt, ectx, sudo, bdy)
			if diagd.HasErrors() {
				return nil, diagd
			}

			if ctxd == nil {
				return nil, terrors.Errorf("unable to find instance loc %q", verr.InstanceLocation)
			}

			diag := &hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     verr.Message,
				Detail:      verr.DetailedOutput().KeywordLocation,
				Subject:     ctxd.Range().Ptr(),
				Expression:  ctxd,
				EvalContext: ectx,
			}

			diags = append(diags, diag)
		}

		// return append(vers, validationErr), nil
	}

	return diags, nil
}

func InstanceLocationStringToHCLRange(instLoc string, msg string, cnt hclsyntax.Expression, ectx *hcl.EvalContext, sudo *SudoContext, file *BodyBuilder) (hcl.Expression, hcl.Diagnostics) {
	splt := strings.Split(strings.TrimPrefix(instLoc, "/"), "/")

	cmp := regexp.MustCompile("additionalProperties '(.*)' not allowed")
	matches := cmp.FindStringSubmatch(msg)
	if len(matches) == 2 {
		splt = append(splt, matches[1])
	}
	// fmt.Println(splt)
	// ctxd := sudo.IdentifyChild(splt...)

	// return ctxd, hcl.Diagnostics{}

	return roll2(splt, cnt, ectx, file)
}

func roll2(splt []string, e hcl.Expression, ectx *hcl.EvalContext, file *BodyBuilder) (hcl.Expression, hcl.Diagnostics) {
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
				return ex, nil

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

		bdy, err := file.NewRootForFile(x.SrcRange.Filename)
		if err != nil {
			return nil, hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid expression",
					Detail:   err.Error(),
					Subject:  e.Range().Ptr(),
				},
			}
		}
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
				for zz, k := range blk.Body.Attributes {
					if zz == splt[0] {
						if len(splt) == 1 {
							return k.Expr, nil
						}
						return roll2(splt[1:], k.Expr, ectx, file)
					}
				}
			}
		}
		// }
	}

	return e, nil
}

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
