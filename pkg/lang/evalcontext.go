package lang

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

func NewContextFromFiles(ctx context.Context, fle map[string][]byte) (*hcl.File, *SudoContext, *BodyBuilder, hcl.Diagnostics, error) {

	bodys := make(map[string]*hclsyntax.Body)

	for k, v := range fle {
		hcldata, errd := hclsyntax.ParseConfig(v, k, hcl.InitialPos)
		if errd.HasErrors() {
			return nil, nil, nil, errd, nil
		}

		// will always work
		bdy := hcldata.Body.(*hclsyntax.Body)

		bodys[k] = bdy
	}

	root := &BodyBuilder{files: bodys}

	mectx := &SudoContext{
		Parent:    nil,
		ParentKey: "",
		Map:       make(map[string]*SudoContext),
	}

	diags := mectx.ApplyBody(ctx, root.NewRoot())

	return nil, mectx, root, diags, nil

}

func NewContextFromFile(ctx context.Context, fle []byte, name string) (*hcl.File, *SudoContext, *BodyBuilder, hcl.Diagnostics, error) {
	return NewContextFromFiles(ctx, map[string][]byte{name: fle})
}

const ArrKey = "____arr"

func EvaluateAttr(ctx context.Context, name string, attr hclsyntax.Expression, parentctx *SudoContext) hcl.Diagnostics {
	childctx := parentctx.BuildStaticEvalContextWithFileData(attr.StartRange().Filename)

	switch e := attr.(type) {
	case *hclsyntax.ObjectConsExpr:

		diags := hcl.Diagnostics{}
		for _, v := range e.Items {
			key, diag := v.KeyExpr.Value(childctx)
			if diag.HasErrors() {
				diags = append(diags, diag...)
				continue
			}

			diag = EvaluateAttr(ctx, key.AsString(), v.ValueExpr, parentctx.NewChild(name))
			if diag.HasErrors() {
				diags = append(diags, diag...)
				continue
			}
		}

		return diags

	case *hclsyntax.TupleConsExpr:

		child := parentctx.NewChild(ArrKey)

		diags := hcl.Diagnostics{}

		for i, v := range e.Exprs {
			diag := EvaluateAttr(ctx, fmt.Sprintf("%d", i), v, child)
			diags = append(diags, diag...)
		}

		delete(parentctx.Map, ArrKey)

		parentctx.ApplyKeyVal(name, child.ToValue())

		return diags

	default:

		val, diag := attr.Value(childctx)
		if diag.HasErrors() {
			return diag
		}

		parentctx.ApplyKeyVal(name, val)

	}

	return hcl.Diagnostics{}
}

func ExtractVariables(ctx context.Context, bdy *hclsyntax.Body, parentctx *SudoContext) hcl.Diagnostics {

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

	for ((len(retrys) > 0 || len(retryattrs) > 0) && (len(prevRetrys) > len(retrys) || len(prevAttrRetrys) > len(retryattrs))) || start {

		start = false
		newRetrys := []*hclsyntax.Block{}
		newAttrRetrys := []*attr{}
		diags := hcl.Diagnostics{}

		for v := range len(retryattrs) + len(retrys) {
			if v < len(retryattrs) {
				attr := retryattrs[v]
				diagd := EvaluateAttr(ctx, attr.Name, attr.Attri.Expr, parentctx)
				if diagd.HasErrors() {
					diags = append(diags, diagd...)
					newAttrRetrys = append(newAttrRetrys, attr)
				}
			} else {
				attr := retrys[v-len(retryattrs)]
				var save *hclsyntax.Attribute
				if attr.Type == "gen" {
					save = attr.Body.Attributes["data"]
					delete(attr.Body.Attributes, "data")
				}
				diagd := NewUnknownBlockEvaluation(ctx, parentctx, attr)
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

		lastDiags = diags
		prevRetrys = retrys
		prevAttrRetrys = retryattrs
		retryattrs = newAttrRetrys
		retrys = newRetrys
	}
	return lastDiags

	// if len(lastDiags) > 0 {
	// 	return lastDiags
	// }

	// isFileParent := parentctx.Parent != nil && parentctx.Parent.ParentKey == FilesKey

	// if isFileParent {

	// }

}

const MetaKey = "____meta"
const FilesKey = "____files"

func NewUnknownBlockEvaluation(ctx context.Context, parentctx *SudoContext, block *hclsyntax.Block) (diags hcl.Diagnostics) {

	strs := []string{block.Type}
	strs = append(strs, block.Labels...)

	child := parentctx.NewNestedChild(strs...)

	userfuncs, _, diag := userfunc.DecodeUserFunctions(block.Body, "func", child.BuildStaticEvalContext)
	if diag.HasErrors() {
		return diag
	}

	child.UserFuncs = userfuncs

	diag = ExtractVariables(ctx, block.Body, child)
	if diag.HasErrors() {
		return diag
	}

	meta := map[string]cty.Value{
		"label": cty.StringVal(strings.Join(block.Labels, ".")),
	}

	if block.Type == "gen" {
		meta["source"] = cty.StringVal(block.TypeRange.Filename)
		meta["root_relative_path"] = cty.StringVal(sanatizeGenPath(child.Map["path"].Value.AsString()))
	}

	meta["block_type"] = cty.StringVal(block.Type)
	meta["done"] = cty.BoolVal(true)

	child.ApplyKeyVal(MetaKey, cty.ObjectVal(meta))

	return hcl.Diagnostics{}

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

func sanitizeFileName(str string) string {
	if !strings.HasPrefix(str, ".retab/") {
		str = ".retab/" + str
	}

	if !strings.HasSuffix(str, ".retab") {
		str = str + ".retab"
	}

	return str
}

func Mapd(ectx map[string]cty.Value, file string, name string) (map[string]cty.Value, error) {
	data := ectx[FilesKey].AsValueMap()[file].AsValueMap()

	if data == nil {
		return nil, terrors.Errorf("file %s not found", file)
	}

	if name != "" {
		data = data[name].AsValueMap()

		if data == nil {
			return nil, terrors.Errorf("block %s not found", name)
		}
	}

	mapper := make(map[string]cty.Value)
	for k, v := range data {
		d := v.AsValueMap()
		if d[MetaKey] == cty.NilVal || d[MetaKey].AsValueMap()["block_type"] == cty.NilVal {
			return nil, terrors.Errorf("block %s:%s has no meta", name, k)
		}
		mapper[k] = v
	}

	return mapper, nil
}

func MapdFile(ectx map[string]cty.Value, file string) (map[string]cty.Value, error) {
	data := ectx[FilesKey].AsValueMap()[file].AsValueMap()

	if data == nil {
		return nil, terrors.Errorf("file %s not found", file)
	}

	if data[MetaKey] == cty.NilVal || data[MetaKey].AsValueMap()["block_type"] == cty.NilVal {
		return nil, terrors.Errorf("file block %s has no meta", file)
	}

	// mapper := make(map[string]cty.Value)
	// for k, v := range data {
	// 	d := v.AsValueMap()
	// 	if d[MetaKey] == cty.NilVal || d[MetaKey].AsValueMap()["block_type"] == cty.NilVal {
	// 		return nil, terrors.Errorf("block %s:%s has no meta", name, k)
	// 	}
	// 	mapper[k] = v
	// }

	return data, nil
}

func NewContextualizedFunctionMap(ectx map[string]cty.Value, file string) map[string]function.Function {

	mapp := function.New(&function.Spec{
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

			resp, err := Mapd(ectx, file, args[0].AsString())
			if err != nil {
				return cty.NilVal, err
			}

			mapd := make(map[string]cty.Value)
			for k, v := range resp {
				mapd[k] = v
			}

			return cty.ObjectVal(mapd), nil
		},
	})

	list := function.New(&function.Spec{
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

			resp, err := Mapd(ectx, file, args[0].AsString())
			if err != nil {
				return cty.NilVal, err
			}

			vals := make([]cty.Value, 0, len(resp))
			for _, v := range resp {
				vals = append(vals, v)
			}

			return cty.TupleVal(vals), nil
		}},
	)

	filed := function.New(&function.Spec{
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
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {

			mapper, err := MapdFile(ectx, sanitizeFileName(args[0].AsString()))
			if err != nil {
				return cty.NilVal, err
			}

			return cty.ObjectVal(mapper), nil
		},
	})

	return map[string]function.Function{
		"file":       filed,
		"allof":      mapp,
		"alloflist":  list,
		"allofarray": list,
	}
}

func NewDynamicContextualizedFunctionMap(ectx *SudoContext) map[string]function.Function {
	// takes in some negative number and returns the nested parent -x levels
	selfer := function.New(&function.Spec{
		Description: `Returns the parent block of the current block`,
		Params:      []function.Parameter{
			// {
			// 	Name:             "levels",
			// 	Type:             cty.Number,
			// 	AllowUnknown:     false,
			// 	AllowDynamicType: false,
			// 	AllowNull:        true,
			// 	Description:      "The number of levels to go up",
			// },
		},
		VarParam: &function.Parameter{
			Name: "levels",
			Type: cty.Number,
		},
		Type: function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {

			if len(args) != 1 {
				// default to 0
				args = append(args, cty.NumberIntVal(0))
			}

			num := args[0].AsBigFloat()
			if !num.IsInt() {
				return cty.NilVal, terrors.Errorf("expected int, got %s", args[0].GoString())
			}

			count, _ := num.Int64()

			if count > 0 {
				return cty.NilVal, terrors.Errorf("expected negative int, got %s", args[0].GoString())
			}

			wrk := ectx

			for range count * -1 {
				if wrk.Parent == nil {
					return cty.NilVal, nil
				}
				wrk = wrk.Parent
			}

			return wrk.ToValue(), nil
		},
	})

	return map[string]function.Function{
		"self": selfer,
	}
}
