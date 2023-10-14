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

	custvars := map[string]map[string]cty.Value{}

	for _, v := range bdy.Blocks {
		if v.Type == "file" {
			continue
		}

		if _, ok := custvars[v.Type]; !ok {
			custvars[v.Type] = map[string]cty.Value{}
		}

		if len(v.Labels) != 1 {
			return nil, nil, nil, errors.Errorf("expected exactly one label, got %d", len(v.Labels))
		}

		mapper := map[string]cty.Value{}

		for _, attr := range v.Body.Attributes {
			val, diag := attr.Expr.Value(ectx)
			if diag.HasErrors() {
				return nil, nil, nil, errors.Wrapf(diag, "failed to evaluate %q", attr.Name)
			}
			mapper[attr.Name] = val
		}

		custvars[v.Type][v.Labels[0]] = cty.ObjectVal(mapper)

	}

	for k, v := range custvars {
		ectx.Variables[k] = cty.ObjectVal(v)
	}

	bdy.Attributes = nil

	return hcldata, ectx, bdy, nil

}
