package hclread

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/go-faster/errors"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/userfunc"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
	"github.com/walteh/terrors"
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

	proc := func(blk *hclsyntax.Block) error {
		key, blks, err := NewAnyBlockEvaluation(ctx, ectx, blk)
		if err != nil {
			return err
		}

		if combos[key] == nil {
			combos[key] = make([]cty.Value, 0)
		}

		combos[key] = append(combos[key], blks)
		return nil
	}

	comp := func() error {
		for k, v := range combos {
			// fmt.Println(k, v)

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
			ectx.Variables[k] = v
		}

		return nil
	}

	retrys := bdy.Blocks
	prevRetrys := []*hclsyntax.Block{}
	start := true
	// starts := 0
	for (len(retrys) > 0 && len(prevRetrys) > len(retrys)) || start {

		start = false
		newRetrys := []*hclsyntax.Block{}

		for _, v := range retrys {
			if v.Type == "file" {
				continue
			}

			err := proc(v)
			if err != nil {
				fmt.Println(err)
				newRetrys = append(newRetrys, v)
			}
		}

		err = comp()
		if err != nil {
			return nil, nil, nil, terrors.Wrapf(err, "failed to combine")
		}

		prevRetrys = retrys
		retrys = newRetrys

		// if len(retrys) == len(prevRetrys) && starts < 3 {
		// 	slices.Reverse(retrys)
		// 	start = true
		// 	starts++
		// }
	}

	// pp.Println(bdy.Blocks)

	// bdy.Attributes = nil

	return hcldata, ectx, bdy, nil

}
