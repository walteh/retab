package hclread

import (
	"context"
	"encoding/base64"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/userfunc"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/walteh/terrors"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

type SudoContext struct {
	ParentKey string
	Parent    *SudoContext
	Map       map[string]*SudoContext
	Array     []*SudoContext
	Value     *cty.Value
	UserFuncs map[string]function.Function
}

func (me *SudoContext) ApplyValue(val cty.Value) {
	if val.Type().IsObjectType() {
		me.ApplyValueMap(val.AsValueMap())
	} else {
		me.Value = &val
		// me.resolved = true
	}
}

func (me *SudoContext) ApplyKeyVal(key string, val cty.Value) {
	// if me.ParentKey == ArrKey {
	// 	me.Parent.ApplyValue(me.ToValue())
	// } else {
	me.NewChild(key).ApplyValue(val)
	// }
}

func (me *SudoContext) ApplyValueMap(val map[string]cty.Value) {
	for k, v := range val {
		me.ApplyKeyVal(k, v)
	}
}

func (me *SudoContext) ApplyArray(arr []cty.Value) {
	if me.ParentKey != ArrKey {
		me.ApplyValue(cty.TupleVal(arr))
		return
	}
	for _, v := range arr {
		me.Array = append(me.Array, &SudoContext{
			Parent: me,
			Value:  &v,
		})
	}
}

func (me *SudoContext) ToValue() cty.Value {
	if me.Value != nil {
		return *me.Value
	}

	if me.ParentKey == ArrKey {
		vars := me.BuildStaticVarsList()
		return cty.TupleVal(vars)
	}

	if me.Array != nil {
		arr := make([]cty.Value, len(me.Array))
		for i, v := range me.Array {
			arr[i] = v.ToValue()
		}
		return cty.TupleVal(arr)
	}

	obj := make(map[string]cty.Value, len(me.Map))
	for k, v := range me.Map {
		obj[k] = v.ToValue()
	}
	return cty.ObjectVal(obj)
}

func (me *SudoContext) BuildStaticEvalVars() map[string]cty.Value {

	wrk := map[string]cty.Value{}
	for k, v := range me.Map {
		wrk[k] = v.ToValue()
	}

	return wrk

}

func (me *SudoContext) BuildStaticVarsList() []cty.Value {

	type sorter struct {
		Key   string
		Value cty.Value
	}

	wrk := make([]sorter, 0, len(me.Map))
	for srt, v := range me.Map {
		if v.Value != nil {
			wrk = append(wrk, sorter{
				Key:   srt,
				Value: *v.Value,
			})
		} else {
			wrk = append(wrk, sorter{
				Key:   srt,
				Value: cty.ObjectVal(v.BuildStaticEvalVars()),
			})
		}
	}

	slices.SortFunc(wrk, func(a, b sorter) int {
		intr, _ := strconv.Atoi(a.Key)
		intr2, _ := strconv.Atoi(b.Key)
		return intr - intr2
	})

	vals := make([]cty.Value, len(wrk))
	for i, v := range wrk {
		vals[i] = v.Value
	}

	return vals

}

func (me *SudoContext) Functions() map[string]function.Function {
	fn := NewFunctionMap()

	for k, v := range NewGlobalContextualizedFunctionMap(me.Root().BuildStaticEvalVars()) {
		fn[k] = v
	}

	if me.UserFuncs != nil {
		for k, v := range me.UserFuncs {
			fn[k] = v
		}
	}

	return fn
}

func (me *SudoContext) NewNestedChild(key ...string) *SudoContext {
	wrk := me
	for _, v := range key {
		wrk = wrk.NewChild(v)
	}
	return wrk
}

func (me *SudoContext) NewArrayChild() *SudoContext {
	return me.NewChild(ArrKey)
}

func (me *SudoContext) NewChild(key string) *SudoContext {

	if me.Map[key] != nil {
		return me.Map[key]
	}

	build := &SudoContext{
		ParentKey: key,
		Parent:    me,
		Map:       make(map[string]*SudoContext),
	}

	me.Map[key] = build

	return build
}

func (wc *SudoContext) BuildStaticEvalContext() *hcl.EvalContext {
	wrk := &hcl.EvalContext{
		Functions: wc.Functions(),
		Variables: wc.BuildStaticEvalVars(),
	}

	return wrk
}

func (wc *SudoContext) BuildStaticEvalContextWithFileData(file string) *hcl.EvalContext {

	internalParent := wc.Root().BuildStaticEvalContext().Variables[FilesKey].AsValueMap()

	if internalParent == nil {
		internalParent = map[string]cty.Value{}

	}

	if internalParent[file] == cty.NilVal {
		internalParent[file] = cty.ObjectVal(map[string]cty.Value{})
	}

	internalParent = internalParent[file].AsValueMap()

	wrk := &hcl.EvalContext{
		Functions: wc.Functions(),
		Variables: map[string]cty.Value{
			"self": cty.ObjectVal(wc.BuildStaticEvalVars()),
		},
	}

	if wc.Parent != nil {
		wrk.Variables["parent"] = cty.ObjectVal(wc.Parent.BuildStaticEvalVars())
		if wc.Parent.Parent != nil {
			wrk.Variables["grandparent"] = cty.ObjectVal(wc.Parent.Parent.BuildStaticEvalVars())
		}
	}

	for k, v := range internalParent {
		wrk.Variables[k] = v
	}

	for k, v := range NewContextualizedFunctionMap(wc.Root().BuildStaticEvalVars(), file) {
		wrk.Functions[k] = v
	}

	return wrk
}

func (wc *SudoContext) Root() *SudoContext {
	if wc.Parent == nil {
		return wc
	}

	return wc.Parent.Root()
}

func ExtractUserFuncs(ctx context.Context, ibdy hcl.Body, parent *hcl.EvalContext) (map[string]function.Function, hcl.Diagnostics) {
	userfuncs, _, diag := userfunc.DecodeUserFunctions(ibdy, "func", func() *hcl.EvalContext { return parent })
	if diag.HasErrors() {
		return nil, diag
	}

	return userfuncs, nil
}

type PreCtyValue struct {
	Val      cty.Value
	Range    hcl.Range
	Name     string
	Children []*PreCtyValue
}

func NewContextFromFiles(ctx context.Context, fle map[string][]byte, parent *hcl.EvalContext) (*hcl.File, *SudoContext, *hclsyntax.Body, map[string]*hclsyntax.Body, hcl.Diagnostics, error) {

	childctx := parent.NewChild()
	childctx.Variables = map[string]cty.Value{}
	childctx.Functions = map[string]function.Function{}

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

	for k, v := range bodys {
		sudoblock := &hclsyntax.Block{
			Type:   FilesKey,
			Body:   v,
			Labels: []string{k},
		}
		root.Blocks = append(root.Blocks, sudoblock)
	}

	mectx := &SudoContext{
		Parent:    nil,
		ParentKey: "",
		Map:       make(map[string]*SudoContext),
	}

	_, diags, err := NewContextFromBody(ctx, root, mectx)

	if err != nil || diags.HasErrors() {
		return nil, nil, nil, nil, diags, err
	}

	return nil, mectx, root, bodys, diags, nil

}

func NewContextFromFile(ctx context.Context, fle []byte, name string) (*hcl.File, *SudoContext, *hclsyntax.Body, hcl.Diagnostics, error) {
	hcldata, errd := hclsyntax.ParseConfig(fle, name, hcl.InitialPos)
	if errd.HasErrors() {
		return nil, nil, nil, errd, nil
	}

	// will always work
	bdy := hcldata.Body.(*hclsyntax.Body)

	ectx := &SudoContext{
		Parent:    nil,
		ParentKey: "",
		Map:       map[string]*SudoContext{},
	}

	ectx.Map[FilesKey] = &SudoContext{
		Parent:    ectx,
		ParentKey: FilesKey,
		Map:       map[string]*SudoContext{},
	}

	ectx.Map[name] = &SudoContext{
		Parent:    ectx,
		ParentKey: name,
		Map:       map[string]*SudoContext{},
	}

	_, diags, err := NewContextFromBody(ctx, bdy, ectx)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return hcldata, ectx, bdy, diags, nil

}

func NewContextFromBody(ctx context.Context, body *hclsyntax.Body, parent *SudoContext) (*hcl.EvalContext, hcl.Diagnostics, error) {

	childctx := parent.BuildStaticEvalContext().NewChild()
	childctx.Variables = map[string]cty.Value{}
	childctx.Functions = map[string]function.Function{}

	diag := ExtractVariables(ctx, body, parent)
	if diag.HasErrors() {
		return nil, diag, nil
	}

	return childctx, hcl.Diagnostics{}, nil
}

func ApplyFileSpecificToContext(ctx context.Context, tange hcl.Range, child *hcl.EvalContext) {

	nest := []string{}

	rootparent := child.Parent()
	for rootparent.Parent() != nil {
		nest = append(nest, rootparent.Variables[MetaKey].AsValueMap()["label"].AsString())
		rootparent = rootparent.Parent()
	}

	files := rootparent.Variables[FilesKey].AsValueMap()
	if files == nil {
		files = map[string]cty.Value{}
	}

	tmp := rootparent.Variables
	for _, v := range nest {
		tmp = tmp[v].AsValueMap()
	}

	files[tange.Filename] = cty.ObjectVal(child.Variables)

	tmp[tange.Filename] = cty.ObjectVal(files)

}

const ArrKey = "____arr"

func EvaluateAttr(ctx context.Context, name string, attr hclsyntax.Expression, parentctx *SudoContext) hcl.Diagnostics {
	childctx := parentctx.BuildStaticEvalContextWithFileData(attr.StartRange().Filename)

	switch e := attr.(type) {
	case *hclsyntax.ObjectConsExpr:
		for _, v := range e.Items {
			key, diag := v.KeyExpr.Value(childctx)
			if diag.HasErrors() {
				return diag
			}

			diag = EvaluateAttr(ctx, key.AsString(), v.ValueExpr, parentctx.NewChild(name))
			if diag.HasErrors() {
				return diag
			}
		}

	case *hclsyntax.TupleConsExpr:

		child := parentctx.NewChild(ArrKey)

		for i, v := range e.Exprs {
			diag := EvaluateAttr(ctx, fmt.Sprintf("%d", i), v, child)
			if diag.HasErrors() {
				return diag
			}
		}

		delete(parentctx.Map, ArrKey)

		parentctx.ApplyKeyVal(name, child.ToValue())

	default:

		val, diag := attr.Value(childctx)
		if diag.HasErrors() {
			return diag
		}

		parentctx.ApplyKeyVal(name, val)

	}

	return nil
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
	// starts := 0
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
		retryattrs = newAttrRetrys
		retrys = newRetrys
	}

	return lastDiags
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

func NewGlobalContextualizedFunctionMap(ectx map[string]cty.Value) map[string]function.Function {
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

				if files, ok := ectx[FilesKey]; ok {
					if files.IsKnown() {
						if files.Type().IsObjectType() {
							sfilename := sanitizeFileName(args[0].AsString())
							if file, ok := files.AsValueMap()[sfilename]; ok {

								return file, nil
							} else {
								known := []string{}
								for k := range files.AsValueMap() {
									known = append(known, k)
								}
								return cty.NilVal, terrors.Errorf("file %s not found, known files: %s", sfilename, strings.Join(known, ", "))
							}
						}
						return cty.NilVal, terrors.Errorf("files is not an object")
					}
					return cty.NilVal, terrors.Errorf("files is not known")
				}
				return cty.NilVal, terrors.Errorf("files not found in context")
			},
		}),
	}
}
func NewContextualizedFunctionMap(ectx map[string]cty.Value, file string) map[string]function.Function {
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

				for nme, blks := range ectx[FilesKey].AsValueMap()[file].AsValueMap() {
					if nme == args[0].AsString() {
						for k, v := range blks.AsValueMap() {
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

				for nme, blks := range ectx[FilesKey].AsValueMap()[file].AsValueMap() {
					if nme == args[0].AsString() {
						for k, v := range blks.AsValueMap() {
							if len(v.AsValueMap()) == 0 {
								return cty.NilVal, terrors.Errorf("block %s has no attributes", k)
							} else {
								mapd = append(mapd, cty.ObjectVal(v.AsValueMap()))
							}
						}
					}
				}

				return cty.ListVal(mapd), nil
			}}),
	}
}
