package hclread

import (
	"context"
	"encoding/base64"
	"fmt"
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

	reevaluateAttr := func(name string, attr *hclsyntax.Attribute) hcl.Diagnostics {
		val, diag := attr.Expr.Value(eectx)
		if diag.HasErrors() {
			return diag
		}
		eectx.Variables[name] = val

		return nil
	}

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

	updateBlocks := func() {
		for k, v := range combos {
			if eectx.Variables[k] == cty.NilVal {
				eectx.Variables[k] = cty.ObjectVal(map[string]cty.Value{})
			}
			wrk := eectx.Variables[k].AsValueMap()
			if wrk == nil {
				wrk = map[string]cty.Value{}
			}

			for _, v2 := range v {
				for k2, v3 := range v2.AsValueMap() {
					wrk[k2] = v3
				}
			}
			eectx.Variables[k] = cty.ObjectVal(wrk)
			combos[k] = nil
		}

		return
	}

	type attr struct {
		Name  string
		Attri *hclsyntax.Attribute
	}

	retryattrs := []*attr{}
	for k, v := range bdy.Attributes {
		retryattrs = append(retryattrs, &attr{
			Name:  k,
			Attri: v,
		})
	}
	prevAttrRetrys := []*attr{}

	retrys := bdy.Blocks
	prevRetrys := []*hclsyntax.Block{}
	lastDiags := hcl.Diagnostics{}
	start := true
	// starts := 0
	for ((len(retrys) > 0 || len(retryattrs) > 0) && (len(prevRetrys) > len(retrys) || len(prevAttrRetrys) > len(retryattrs))) || start {

		start = false
		newRetrys := []*hclsyntax.Block{}
		newAttrRetrys := []*attr{}
		diags := hcl.Diagnostics{}

		for v := range len(retryattrs) + len(retrys) {
			if v < len(retryattrs) {
				attr := retryattrs[v]
				diagd := reevaluateAttr(attr.Name, attr.Attri)
				if diagd.HasErrors() {
					diags = append(diags, diagd...)
					newAttrRetrys = append(newAttrRetrys, attr)
				}
			} else {
				attr := retrys[v-len(retryattrs)]
				var save *hclsyntax.Attribute
				if attr.Type == "gen" {
					// this could probably be better, its just to skip the processing of the data attribute until the NewGenBlockEvaluation
					// tbh not even sure if this is needed anymore
					save = attr.Body.Attributes["data"]
					// attr.Body.Attributes["data"] = &hclsyntax.Attribute{
					// 	Expr: NewBrokenExpression(hcl.Diagnostics{
					// 		{
					// 			Severity: hcl.DiagError,
					// 			Summary:  "gen data not supported",
					// 			Detail:   "gen blocks are not supported in this context",
					// 			Subject:  attr.TypeRange.Ptr(),
					// 		},
					// 	}),
					// }
					delete(attr.Body.Attributes, "data")
				}
				diagd := reevaluate(attr)
				if save != nil {
					attr.Body.Attributes["data"] = save
				}
				if diagd.HasErrors() {
					diags = append(diags, diagd...)
					newRetrys = append(newRetrys, attr)
				}
			}
		}

		if len(diags) < len(lastDiags) {
			start = true
		}

		updateBlocks()

		lastDiags = diags
		prevRetrys = retrys
		retryattrs = newAttrRetrys
		retrys = newRetrys
	}

	return eectx.Variables, lastDiags
}

const MetaKey = "____meta"
const FilesKey = "____files"

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

	if block.Type == "gen" {
		meta["source"] = cty.StringVal(block.TypeRange.Filename)
		meta["root_relative_path"] = cty.StringVal(sanatizeGenPath(tmp["path"].AsString()))
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

func NewContextFromBody(ctx context.Context, body *hclsyntax.Body, parent *hcl.EvalContext) (*hcl.EvalContext, hcl.Diagnostics, error) {

	ectx := parent.NewChild()

	funcs, diag := ExtractUserFuncs(ctx, body, ectx)
	if diag.HasErrors() {
		return nil, diag, nil
	}

	if ectx.Functions == nil {
		ectx.Functions = map[string]function.Function{}
	}

	for k, v := range funcs {
		ectx.Functions[k] = v
	}

	vars, diag := ExtractVariables(ctx, body, ectx)
	if diag.HasErrors() {
		return nil, diag, nil
	}

	if ectx.Variables == nil {
		ectx.Variables = map[string]cty.Value{}
	}

	for k, v := range vars {
		ectx.Variables[k] = v
	}

	return ectx, hcl.Diagnostics{}, nil
}

func NewContextFromFiles(ctx context.Context, fle map[string][]byte, parent *hcl.EvalContext) (*hcl.File, *hcl.EvalContext, *hclsyntax.Body, map[string]*hclsyntax.Body, hcl.Diagnostics, error) {

	ectx := parent.NewChild()

	bodys := make(map[string]*hclsyntax.Body)

	for k, v := range fle {
		hcldata, errd := hclsyntax.ParseConfig(v, k, hcl.InitialPos)
		if errd.HasErrors() {
			return nil, nil, nil, nil, errd, nil
		}

		// will always work
		bdy := hcldata.Body.(*hclsyntax.Body)

		bodys[k] = bdy
	}

	root := &hclsyntax.Body{
		Attributes: hclsyntax.Attributes{},
		Blocks:     make([]*hclsyntax.Block, 0),
	}

	for _, v := range bodys {
		root.Blocks = append(root.Blocks, v.Blocks...)
		for k, v2 := range v.Attributes {
			root.Attributes[k] = v2
		}
	}

	ccc, diags, err := NewContextFromBody(ctx, root, ectx)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	return nil, ccc, root, bodys, diags, nil

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

	ctxfuncs := NewContextualizedFunctionMap(ectx)
	for k, v := range ctxfuncs {
		ectx.Functions[k] = v
	}

	// will always work
	bdy := hcldata.Body.(*hclsyntax.Body)

	ccc, diags, err := NewContextFromBody(ctx, bdy, ectx)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return hcldata, ccc, bdy, diags, nil

}

func NewGetMetaKeyFunc(str string) function.Function {
	return function.New(&function.Spec{
		Description: fmt.Sprintf(`Gets the meta %s of an hcl block`, str),
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

			return cty.StringVal(mp[str].AsString()), nil
		},
	})
}
func NewFunctionMap() map[string]function.Function {

	return map[string]function.Function{
		"jsonencode":             stdlib.JSONEncodeFunc,
		"jsondecode":             stdlib.JSONDecodeFunc,
		"csvdecode":              stdlib.CSVDecodeFunc,
		"equal":                  stdlib.EqualFunc,
		"notequal":               stdlib.NotEqualFunc,
		"concat":                 stdlib.ConcatFunc,
		"format":                 stdlib.FormatFunc,
		"join":                   stdlib.JoinFunc,
		"merge":                  stdlib.MergeFunc,
		"length":                 stdlib.LengthFunc,
		"keys":                   stdlib.KeysFunc,
		"values":                 stdlib.ValuesFunc,
		"flatten":                stdlib.FlattenFunc,
		"coelesce":               stdlib.CoalesceFunc,
		"contains":               stdlib.ContainsFunc,
		"index":                  stdlib.IndexFunc,
		"lookup":                 stdlib.LookupFunc,
		"element":                stdlib.ElementFunc,
		"slice":                  stdlib.SliceFunc,
		"compact":                stdlib.CompactFunc,
		"distinct":               stdlib.DistinctFunc,
		"reverselist":            stdlib.ReverseListFunc,
		"setproduct":             stdlib.SetProductFunc,
		"setunion":               stdlib.SetUnionFunc,
		"setintersection":        stdlib.SetIntersectionFunc,
		"sethaselement":          stdlib.SetHasElementFunc,
		"setsubtract":            stdlib.SetSubtractFunc,
		"setsymmetricdifference": stdlib.SetSymmetricDifferenceFunc,
		"formatdate":             stdlib.FormatDateFunc,
		"timeadd":                stdlib.TimeAddFunc,
		"add":                    stdlib.AddFunc,
		"assertnotnull":          stdlib.AssertNotNullFunc,
		"byteslen":               stdlib.BytesLenFunc,
		"byteslice":              stdlib.BytesSliceFunc,
		"not":                    stdlib.NotFunc,
		"and":                    stdlib.AndFunc,
		"or":                     stdlib.OrFunc,
		"upper":                  stdlib.UpperFunc,
		"lower":                  stdlib.LowerFunc,
		"replace":                stdlib.ReplaceFunc,
		"split":                  stdlib.SplitFunc,
		"substr":                 stdlib.SubstrFunc,
		"trimprefix":             stdlib.TrimPrefixFunc,
		"trimsuffix":             stdlib.TrimSuffixFunc,
		"trimspace":              stdlib.TrimSpaceFunc,
		"trim":                   stdlib.TrimFunc,
		"chomp":                  stdlib.ChompFunc,
		"chunklist":              stdlib.ChunklistFunc,
		"coalesce":               stdlib.CoalesceFunc,
		"indent":                 stdlib.IndentFunc,
		"title":                  stdlib.TitleFunc,
		"abs":                    stdlib.AbsoluteFunc,
		"ceil":                   stdlib.CeilFunc,
		"div":                    stdlib.DivideFunc,
		"mod":                    stdlib.ModuloFunc,
		"floor":                  stdlib.FloorFunc,
		"max":                    stdlib.MaxFunc,
		"min":                    stdlib.MinFunc,
		"mul":                    stdlib.MultiplyFunc,
		"gte":                    stdlib.GreaterThanOrEqualToFunc,
		"gt":                     stdlib.GreaterThanFunc,
		"lte":                    stdlib.LessThanOrEqualToFunc,
		"lt":                     stdlib.LessThanFunc,
		"sub":                    stdlib.SubtractFunc,
		"neg":                    stdlib.NegateFunc,
		"int":                    stdlib.IntFunc,
		"log":                    stdlib.LogFunc,
		"pow":                    stdlib.PowFunc,
		"signum":                 stdlib.SignumFunc,
		"parseint":               stdlib.ParseIntFunc,
		"range":                  stdlib.RangeFunc,
		"formatlist":             stdlib.FormatListFunc,
		"regex":                  stdlib.RegexFunc,
		"regexall":               stdlib.RegexAllFunc,
		"regexreplace":           stdlib.RegexReplaceFunc,
		"zipmap":                 stdlib.ZipmapFunc,
		"coelscelist":            stdlib.CoalesceListFunc,
		"reverse":                stdlib.ReverseFunc,
		"sort":                   stdlib.SortFunc,
		"ref":                    NewGetMetaKeyFunc("ref"),
		"source":                 NewGetMetaKeyFunc("source"),
		"output":                 NewGetMetaKeyFunc("root_relative_path"),
		"label":                  NewGetMetaKeyFunc("label"),
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

func NewGlobalContextualizedFunctionMap(ectx *hcl.EvalContext) map[string]function.Function {
	return map[string]function.Function{
		"file": function.New(&function.Spec{
			Description: "Returns the contents of another .retab file",
			Params: []function.Parameter{
				{
					Name:             "file",
					Type:             cty.String,
					AllowUnknown:     false,
					AllowDynamicType: false,
					AllowNull:        false,
				},
			},
			Type: function.StaticReturnType(cty.DynamicPseudoType),
			// RefineResult: refineNonNull,
			Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
				if files, ok := ectx.Variables[FilesKey]; ok {
					if files.IsKnown() {
						if files.Type().IsObjectType() {
							if file, ok := files.AsValueMap()[args[0].AsString()]; ok {
								return file, nil
							} else {
								known := []string{}
								for k := range files.AsValueMap() {
									known = append(known, k)
								}
								return cty.NilVal, terrors.Errorf("file %s not found, known files: %s", args[0].AsString(), strings.Join(known, ", "))
							}
						}
					}
				}
				return cty.NilVal, terrors.Errorf("files not found in context")
			},
		}),
	}
}
func NewContextualizedFunctionMap(ectx *hcl.EvalContext) map[string]function.Function {
	return map[string]function.Function{

		"allof": function.New(&function.Spec{
			Description: `Returns a map of all blocks w\ the given label`,
			Params: []function.Parameter{
				{
					Name:             "block",
					Type:             cty.String,
					AllowUnknown:     true,
					AllowDynamicType: true,
					AllowNull:        false,
					AllowMarked:      true,
				},
			},
			Type: function.StaticReturnType(cty.DynamicPseudoType),
			Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {

				mapd := make(map[string]cty.Value)

				for nme, blks := range ectx.Variables {
					if nme == args[0].AsString() {
						objd := blks.AsValueMap()
						for k, v := range objd {
							mapd[k] = v
						}
					}
				}

				return cty.ObjectVal(mapd), nil
			},
		}),
		"alloflist": function.New(&function.Spec{
			Description: `Returns a list of all blocks w\ the given label`,
			Params: []function.Parameter{
				{
					Name:             "block",
					Type:             cty.String,
					AllowUnknown:     true,
					AllowDynamicType: true,
					AllowNull:        false,
					AllowMarked:      true,
				},
			},
			Type: function.StaticReturnType(cty.DynamicPseudoType),
			Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {

				mapd := make([]cty.Value, 0)

				for nme, blks := range ectx.Variables {
					if nme == args[0].AsString() {
						objd := blks.AsValueMap()
						for _, v := range objd {
							mapd = append(mapd, v)
						}
					}
				}

				return cty.ListVal(mapd), nil
			}}),
	}
}

// type BrokenExpression struct {
// 	hclsyntax.Node

// 	// Node

// 	// // The hcl.Expression methods are duplicated here, rather than simply
// 	// // embedded, because both Node and hcl.Expression have a Range method
// 	// // and so they conflict.

// 	// Value(ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics)
// 	// Variables() []hcl.Traversal
// 	// StartRange() hcl.Range
// }

// func NewBrokenExpression() *BrokenExpression {
// 	return &BrokenExpression{}
// }

// var _ hclsyntax.Expression = (*BrokenExpression)(nil)

// func (be *BrokenExpression) Value(ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
// 	val, err := ctx.Functions["jsondecode"].Proxy()(cty.StringVal(`{"error": "broken expression"}`))
// 	if err != nil {
// 		return cty.NilVal, hcl.Diagnostics{
// 			{
// 				Summary:  "broken expression",
// 				Detail:   "broken expression",
// 				Severity: hcl.DiagError,
// 			},
// 		}
// 	}
// 	return val, nil
// }

// func (be *BrokenExpression) Variables() []hcl.Traversal {
// 	return nil
// }

// func (be *BrokenExpression) StartRange() hcl.Range {
// 	return hcl.Range{}
// }
