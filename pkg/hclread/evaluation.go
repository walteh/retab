package hclread

import (
	"context"
	"io"

	"github.com/go-faster/errors"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/userfunc"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

// old implementation, for backwards compatibility
func NewEvaluation(ctx context.Context, fle afero.File) (*hcl.File, *hcl.EvalContext, *hclsyntax.Body, error) {
	defer fle.Close()

	return NewEvaluationReadCloser(ctx, fle, fle.Name())
}

func NewEvaluationReadCloser(ctx context.Context, fle io.Reader, name string) (*hcl.File, *hcl.EvalContext, *hclsyntax.Body, error) {

	all, err := afero.ReadAll(fle)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "failed to read %q", name)
	}

	hcldata, errd := hclsyntax.ParseConfig(all, name, hcl.InitialPos)
	if errd.HasErrors() {
		return nil, nil, nil, errors.Wrapf(errd, "failed to parse %q", name)
	}

	ectx := &hcl.EvalContext{
		Functions: map[string]function.Function{
			"jsonencode": stdlib.JSONEncodeFunc,
			"jsondecode": stdlib.JSONDecodeFunc,
		},
		Variables: map[string]cty.Value{},
	}

	userfuncs, rbdy, diag := userfunc.DecodeUserFunctions(hcldata.Body, "func", func() *hcl.EvalContext { return ectx })
	if diag.HasErrors() {
		return nil, nil, nil, errors.Wrapf(diag, "failed to decode user functions")
	}

	for k, v := range userfuncs {
		ectx.Functions[k] = v
	}

	// this will always work
	bdy := rbdy.(*hclsyntax.Body)

	blks := hclsyntax.Blocks{}
	for _, v := range bdy.Blocks {
		if v.Type == "func" {
			continue
		}
		blks = append(blks, v)
	}

	bdy.Blocks = blks

	// process attributes

	for _, v := range bdy.Attributes {
		val, diag := v.Expr.Value(ectx)
		if diag.HasErrors() {
			return nil, nil, nil, errors.Wrapf(diag, "failed to evaluate %q", v.Name)
		}
		ectx.Variables[v.Name] = val
	}

	custvars := map[string]cty.Value{}

	combos := make(map[string][]cty.Value, 0)

	for _, v := range bdy.Blocks {
		if v.Type == "file" {
			continue
		}

		key, blks, err := NewAnyBlockEvaluation(ctx, ectx, v)
		if err != nil {
			return nil, nil, nil, err
		}

		if combos[key] == nil {
			combos[key] = make([]cty.Value, 0)
		}

		combos[key] = append(combos[key], blks)

	}

	for k, v := range combos {
		for _, v2 := range v {
			if custvars[k] == cty.NilVal {
				custvars[k] = cty.ObjectVal(map[string]cty.Value{})
			}
			wrk := custvars[k].AsValueMap()
			for k2, v3 := range v2.AsValueMap() {
				if wrk == nil {
					wrk = map[string]cty.Value{}
				}

				wrk[k2] = v3
			}
			custvars[k] = cty.ObjectVal(wrk)
		}
	}

	for k, v := range custvars {

		ectx.Variables[k] = v
	}

	bdy.Attributes = nil

	return hcldata, ectx, bdy, nil

}
