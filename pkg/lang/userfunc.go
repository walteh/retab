package lang

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

var funcBodySchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "params",
			Required: true,
		},
		{
			Name:     "variadic_param",
			Required: false,
		},
		{
			Name:     "result",
			Required: true,
		},
	},
}

type ContextFunc func() *hcl.EvalContext

func decodeUserFunctions(tctx context.Context, body hcl.Body, blockType string, contextFunc ContextFunc, parentctx *SudoContext) (funcs map[string]function.Function, remain hcl.Body, diags hcl.Diagnostics) {
	schema := &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       blockType,
				LabelNames: []string{"name"},
			},
		},
	}

	content, remain, diags := body.PartialContent(schema)
	if diags.HasErrors() {
		return nil, remain, diags
	}

	// first call to getBaseCtx will populate context, and then the same
	// context will be used for all subsequent calls. It's assumed that
	// all functions in a given body should see an identical context.
	// var baseCtx *hcl.EvalContext
	// getBaseCtx := func() *hcl.EvalContext {
	// 	if baseCtx == nil {
	// 		if contextFunc != nil {
	// 			baseCtx = contextFunc()
	// 		}
	// 	}
	// 	// baseCtx might still be nil here, and that's okay
	// 	return baseCtx
	// }

	funcs = make(map[string]function.Function)
Blocks:
	for _, block := range content.Blocks {
		name := block.Labels[0]
		funcContent, funcDiags := block.Body.Content(funcBodySchema)
		diags = append(diags, funcDiags...)
		if funcDiags.HasErrors() {
			continue
		}

		paramsExpr := funcContent.Attributes["params"].Expr
		resultExpr := funcContent.Attributes["result"]
		var varParamExpr hcl.Expression
		if funcContent.Attributes["variadic_param"] != nil {
			varParamExpr = funcContent.Attributes["variadic_param"].Expr
		}

		var params []string
		var varParam string

		paramExprs, paramsDiags := hcl.ExprList(paramsExpr)
		diags = append(diags, paramsDiags...)
		if paramsDiags.HasErrors() {
			continue
		}
		for _, paramExpr := range paramExprs {
			param := hcl.ExprAsKeyword(paramExpr)
			if param == "" {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid param element",
					Detail:   "Each parameter name must be an identifier.",
					Subject:  paramExpr.Range().Ptr(),
				})
				continue Blocks
			}
			params = append(params, param)
		}

		if varParamExpr != nil {
			varParam = hcl.ExprAsKeyword(varParamExpr)
			if varParam == "" {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid variadic_param",
					Detail:   "The variadic parameter name must be an identifier.",
					Subject:  varParamExpr.Range().Ptr(),
				})
				continue
			}
		}

		spec := &function.Spec{}
		for _, paramName := range params {
			spec.Params = append(spec.Params, function.Parameter{
				Name: paramName,
				Type: cty.DynamicPseudoType,
			})
		}
		if varParamExpr != nil {
			spec.VarParam = &function.Parameter{
				Name: varParam,
				Type: cty.DynamicPseudoType,
			}
		}
		impl := func(args []cty.Value) (cty.Value, error) {
			ctx := &hcl.EvalContext{
				Variables: make(map[string]cty.Value),
				Functions: map[string]function.Function{},
			}
			ctx.Variables = make(map[string]cty.Value)

			// The cty function machinery guarantees that we have at least
			// enough args to fill all of our params.
			for i, paramName := range params {
				ctx.Variables[paramName] = args[i]
			}
			if spec.VarParam != nil {
				varArgs := args[len(params):]
				ctx.Variables[varParam] = cty.TupleVal(varArgs)
			}

			child, err := parentctx.NewNonBlockChild(resultExpr.Name, resultExpr.NameRange)
			if err != nil {
				return cty.DynamicVal, hcl.Diagnostics{
					{
						Severity: hcl.DiagError,
						Summary:  "block already exists with this name",
						Detail:   err.Error(),
						Subject:  &resultExpr.NameRange,
					},
				}
			}

			delete(parentctx.Map, resultExpr.Name)

			attr := NewObjectItemAttribute("___check", resultExpr.Range, resultExpr.Expr.(hclsyntax.Expression))

			diags = EvaluateAttr(tctx, attr, child, ctx)
			if diags.HasErrors() {
				// Smuggle the diagnostics out via the error channel, since
				// a diagnostics sequence implements error. Caller can
				// type-assert this to recover the individual diagnostics
				// if desired.
				return cty.DynamicVal, diags
			}

			// get the value of the block at ___check

			return child.Map["___check"].ToValueWithExtraContext(), nil
		}
		spec.Type = func(args []cty.Value) (cty.Type, error) {
			val, err := impl(args)
			return val.Type(), err
		}
		spec.Impl = func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			return impl(args)
		}
		funcs[name] = function.New(spec)
	}

	return funcs, remain, diags
}
