package hclread

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/userfunc"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

func NewEvaluation(ctx context.Context, fle afero.File) (*hcl.EvalContext, *hclsyntax.Body, error) {
	defer fle.Close()

	all, err := afero.ReadAll(fle)
	if err != nil {
		return nil, nil, err
	}

	hcldata, errd := hclsyntax.ParseConfig(all, fle.Name(), hcl.InitialPos)
	if errd.HasErrors() {
		return nil, nil, errd
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
		return nil, nil, diag
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
			return nil, nil, diag
		}
		ectx.Variables[v.Name] = val
	}

	bdy.Attributes = nil

	return ectx, bdy, nil

}
