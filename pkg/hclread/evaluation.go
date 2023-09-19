package hclread

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

type Evalutaion struct {
	ectx *hcl.EvalContext
	Body *hclsyntax.Body
}

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

	// this will always work
	bdy := hcldata.Body.(*hclsyntax.Body)

	ectx := &hcl.EvalContext{
		Functions: map[string]function.Function{
			"jsonencode": stdlib.JSONEncodeFunc,
			"jsondecode": stdlib.JSONDecodeFunc,
		},
		Variables: map[string]cty.Value{},
	}

	for _, v := range bdy.Attributes {
		val, diag := v.Expr.Value(ectx)
		if diag.HasErrors() {
			return nil, nil, diag
		}
		ectx.Variables[v.Name] = val
	}

	return ectx, bdy, nil

}
