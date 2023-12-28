package hclread

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/userfunc"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/walteh/terrors"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

func ExtractUserFuncs(ctx context.Context, ibdy hcl.Body, parent *hcl.EvalContext) (map[string]function.Function, hcl.Diagnostics) {
	userfuncs, _, diag := userfunc.DecodeUserFunctions(ibdy, "func", func() *hcl.EvalContext { return parent })
	if diag.HasErrors() {
		return nil, diag
	}

	return userfuncs, nil
}

func ExtractVariables(ctx context.Context, bdy *hclsyntax.Body, parent *hcl.EvalContext) (map[string]cty.Value, hcl.Diagnostics) {

	eectx := parent.NewChild()

	eectx.Variables = map[string]cty.Value{}

	for _, v := range bdy.Attributes {
		val, diag := v.Expr.Value(eectx)
		if diag.HasErrors() {
			return nil, diag
		}
		eectx.Variables[v.Name] = val
	}

	custvars := map[string]cty.Value{}

	combos := make(map[string][]cty.Value, 0)

	reevaluate := func(blk *hclsyntax.Block) hcl.Diagnostics {
		key, blks, diags := NewUnknownBlockEvaluation(ctx, eectx, blk)
		if diags.HasErrors() {
			return diags
		}

		if combos[key] == nil {
			combos[key] = make([]cty.Value, 0)
		}

		combos[key] = append(combos[key], blks)
		return nil
	}

	updateVariables := func() {
		for k, v := range combos {
			if custvars[k] == cty.NilVal {
				custvars[k] = cty.ObjectVal(map[string]cty.Value{})
			}
			wrk := custvars[k].AsValueMap()
			if wrk == nil {
				wrk = map[string]cty.Value{}
			}

			for _, v2 := range v {
				for k2, v3 := range v2.AsValueMap() {
					wrk[k2] = v3
				}
			}
			custvars[k] = cty.ObjectVal(wrk)
			combos[k] = nil
		}

		for k, v := range custvars {
			eectx.Variables[k] = v
		}

		return
	}

	retrys := bdy.Blocks
	prevRetrys := []*hclsyntax.Block{}
	lastDiags := hcl.Diagnostics{}
	start := true
	// starts := 0
	for (len(retrys) > 0 && len(prevRetrys) > len(retrys)) || start {

		start = false
		newRetrys := []*hclsyntax.Block{}

		for _, v := range retrys {
			if v.Type == "gen" {
				continue
			}

			diags := reevaluate(v)
			if diags.HasErrors() {
				lastDiags = diags
				newRetrys = append(newRetrys, v)
			}
		}

		updateVariables()

		prevRetrys = retrys
		retrys = newRetrys
	}

	return eectx.Variables, lastDiags

}

const MetaKey = "____meta"

func NewUnknownBlockEvaluation(ctx context.Context, ectx *hcl.EvalContext, block *hclsyntax.Block) (key string, res cty.Value, diags hcl.Diagnostics) {

	tmp := make(map[string]cty.Value)

	for _, attr := range block.Body.Attributes {
		// Evaluate the attribute's expression to get a cty.Value
		val, err := attr.Expr.Value(ectx)
		if err.HasErrors() {
			return "", cty.Value{}, err
		}

		tmp[attr.Name] = val
	}

	meta := map[string]cty.Value{
		"label": cty.StringVal(strings.Join(block.Labels, ".")),
	}

	tmp[MetaKey] = cty.ObjectVal(meta)

	for _, blkd := range block.Body.Blocks {

		key, blks, diags := NewUnknownBlockEvaluation(ctx, ectx, blkd)
		if diags.HasErrors() {
			return "", cty.Value{}, diags
		}

		if tmp[key] == cty.NilVal {
			tmp[key] = cty.ObjectVal(map[string]cty.Value{})
		}

		wrk := tmp[key].AsValueMap()
		if wrk == nil {
			wrk = map[string]cty.Value{}
		}

		for k, v := range blks.AsValueMap() {
			wrk[k] = v
		}

		tmp[key] = cty.ObjectVal(wrk)
	}

	for _, lab := range block.Labels {
		tmp = map[string]cty.Value{
			lab: cty.ObjectVal(tmp),
		}
	}

	return block.Type, cty.ObjectVal(tmp), hcl.Diagnostics{}

}

func NewContextFromFile(ctx context.Context, fle []byte, name string) (*hcl.File, *hcl.EvalContext, *hclsyntax.Body, hcl.Diagnostics, error) {

	hcldata, errd := hclsyntax.ParseConfig(fle, name, hcl.InitialPos)
	if errd.HasErrors() {
		return nil, nil, nil, errd, nil
	}

	ectx := &hcl.EvalContext{
		Functions: NewFunctionMap(),
		Variables: map[string]cty.Value{},
	}

	// will always work
	bdy := hcldata.Body.(*hclsyntax.Body)

	// process funcs
	funcs, diag := ExtractUserFuncs(ctx, bdy, ectx)
	if diag.HasErrors() {
		return nil, nil, nil, diag, nil
	}

	for k, v := range funcs {
		ectx.Functions[k] = v
	}

	// todo, do we need to remove the func blocks from the body?

	// process variables
	vars, diag := ExtractVariables(ctx, bdy, ectx)
	if diag.HasErrors() {
		return nil, nil, nil, diag, nil
	}

	for k, v := range vars {
		ectx.Variables[k] = v
	}

	return hcldata, ectx, bdy, nil, nil
}

type WorkingContext struct {
	ectx *hcl.EvalContext
}

func (me *WorkingContext) EvalContext() *hcl.EvalContext { return me.ectx }

func NewFunctionMap() map[string]function.Function {

	return map[string]function.Function{
		"jsonencode": stdlib.JSONEncodeFunc,
		"jsondecode": stdlib.JSONDecodeFunc,
		"csvdecode":  stdlib.CSVDecodeFunc,
		// "yamlencode": stdlib.YAMLDecodeFunc,
		"equal":      stdlib.EqualFunc,
		"notequal":   stdlib.NotEqualFunc,
		"concat":     stdlib.ConcatFunc,
		"format":     stdlib.FormatFunc,
		"join":       stdlib.JoinFunc,
		"lower":      stdlib.LowerFunc,
		"upper":      stdlib.UpperFunc,
		"replace":    stdlib.ReplaceFunc,
		"split":      stdlib.SplitFunc,
		"substr":     stdlib.SubstrFunc,
		"trimprefix": stdlib.TrimPrefixFunc,
		"trimspace":  stdlib.TrimSpaceFunc,
		"trimsuffix": stdlib.TrimSuffixFunc,
		"chomp":      stdlib.ChompFunc,
		"label": function.New(&function.Spec{
			Description: `Gets the label of an hcl block`,
			Params: []function.Parameter{
				{
					Name:             "block",
					Type:             cty.DynamicPseudoType,
					AllowUnknown:     true,
					AllowDynamicType: true,
					AllowNull:        false,
					AllowMarked:      true,
				},
			},
			Type: function.StaticReturnType(cty.String),
			Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
				if len(args) != 1 {
					return cty.NilVal, terrors.Errorf("expected 1 argument, got %d", len(args))
				}

				mp := args[0].AsValueMap()
				if mp == nil {
					return cty.NilVal, terrors.Errorf("expected map, got %s", args[0].GoString())
				}

				if mp[MetaKey] == cty.NilVal {
					return cty.NilVal, terrors.Errorf("expected map with _label, got %s", args[0].GoString())
				}

				mp = mp[MetaKey].AsValueMap()
				if mp == nil {
					return cty.NilVal, terrors.Errorf("expected map with _label, got %s", args[0].GoString())
				}

				return cty.StringVal(mp["label"].AsString()), nil
			},
		}),
		"base64encode": function.New(&function.Spec{
			Description: `Returns the Base64-encoded version of the given string.`,
			Params: []function.Parameter{
				{
					Name:             "str",
					Type:             cty.String,
					AllowUnknown:     false,
					AllowDynamicType: false,
					AllowNull:        false,
				},
			},
			Type: function.StaticReturnType(cty.String),
			// RefineResult: refineNonNull,
			Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
				if len(args) != 1 {
					return cty.NilVal, terrors.Errorf("expected 1 argument, got %d", len(args))
				}
				if args[0].IsNull() {
					return cty.StringVal(""), nil
				}

				if args[0].Type() != cty.String {
					return cty.NilVal, terrors.Errorf("expected string, got %s", args[0].GoString())
				}
				return cty.StringVal(base64.StdEncoding.EncodeToString([]byte(args[0].AsString()))), nil
			},
		}),
		"base64decode": function.New(&function.Spec{
			Description: `Returns the Base64-decoded version of the given string.`,
			Params: []function.Parameter{
				{
					Name:             "str",
					Type:             cty.String,
					AllowUnknown:     false,
					AllowDynamicType: false,
					AllowNull:        false,
				},
			},
			Type: function.StaticReturnType(cty.String),
			// RefineResult: refineNonNull,
			Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
				if len(args) != 1 {
					return cty.NilVal, terrors.Errorf("expected 1 argument, got %d", len(args))
				}
				if args[0].IsNull() {
					return cty.StringVal(""), nil
				}
				if args[0].Type() != cty.String {
					return cty.NilVal, terrors.Errorf("expected string, got %s", args[0].GoString())
				}
				dec, err := base64.StdEncoding.DecodeString(args[0].AsString())
				if err != nil {
					return cty.NilVal, err
				}
				return cty.StringVal(string(dec)), nil
			},
		}),
	}
}
