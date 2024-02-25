package lang

import (
	"context"
	"encoding/base64"
	"fmt"
	"path/filepath"
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

		bodys[sanitizeFileName(k)] = bdy
	}

	root := &BodyBuilder{files: bodys}

	mectx := &SudoContext{
		Parent:           nil,
		ParentKey:        "",
		Map:              make(map[string]*SudoContext),
		UserFuncs:        map[string]function.Function{},
		Meta:             &SimpleNameMeta{hcl.Range{}},
		TmpFileLevelVars: map[string]cty.Value{},
		isArray:          false,
	}

	diags := mectx.ApplyBody(ctx, root.NewRoot())

	return nil, mectx, root, diags, nil

}

func NewContextFromFile(ctx context.Context, fle []byte, name string) (*hcl.File, *SudoContext, *BodyBuilder, hcl.Diagnostics, error) {
	return NewContextFromFiles(ctx, map[string][]byte{name: fle})
}

const ArrKey = "____arr"

const FuncKey = "____func"

func EvaluateAttr(ctx context.Context, attr *hclsyntax.Attribute, parentctx *SudoContext, extra ...*hcl.EvalContext) hcl.Diagnostics {
	childctx := parentctx.BuildStaticEvalContextWithFileData(attr.NameRange.Filename)

	for _, v := range extra {
		for k, v := range v.Variables {
			childctx.Variables[k] = v
		}
		for k, v := range v.Functions {
			childctx.Functions[k] = v
		}
	}

	switch e := attr.Expr.(type) {
	case *hclsyntax.ObjectConsExpr:

		child := parentctx.NewChild(attr.Name, attr.NameRange)

		diags := hcl.Diagnostics{}
		for _, v := range e.Items {
			key, diag := v.KeyExpr.Value(childctx)
			key, _ = key.Unmark()
			if diag.HasErrors() {
				diags = append(diags, diag...)
				continue
			}

			attrn := NewObjectItemAttribute(key.AsString(), v.KeyExpr.Range(), v.ValueExpr)
			attrn.NameRange = v.ValueExpr.Range()

			diag = EvaluateAttr(ctx, attrn, child)
			if diag.HasErrors() {
				diags = append(diags, diag...)
			}

			if len(diags) == 0 {

				child.Map[key.AsString()].Meta = &SimpleNameMeta{v.KeyExpr.Range()}
			}
		}

		return diags

	case *hclsyntax.TupleConsExpr:

		child := parentctx.NewChild(attr.Name, attr.NameRange)
		child.isArray = true

		diags := hcl.Diagnostics{}

		for i, v := range e.Exprs {
			attrn := NewArrayItemAttribute(i, v)
			attrn.NameRange = v.Range()
			diag := EvaluateAttr(ctx, attrn, child)
			diags = append(diags, diag...)
		}

		if len(diags) > 0 {
			return diags
		}

		// child.Meta = &SimpleNameMeta{attr.NameRange}

		// // we don't keep the array in its normal state, we convert it to a real array
		// delete(parentctx.Map, ArrKey)

		// parentctx.ApplyArray(child.List())

		return diags
	// case *hclsyntax.TemplateExpr:
	// 	diags := hcl.Diagnostics{}

	// 	child := parentctx.NewChild(FuncKey)

	// 	for _, v := range e.Parts {
	// 		diag := EvaluateAttr(ctx, name, v, parentctx)
	// 		diags = append(diags, diag...)
	// 	}

	// 	delete(parentctx.Map, FuncKey)

	// 	if len(diags) > 0 {
	// 		return diags
	// 	}

	// 	hi := child.ToValue()

	// 	parentctx.ApplyKeyVal(name, hi)

	case *hclsyntax.ForExpr:

		if _, ok := e.CollExpr.(*PreCalcExpr); ok {
			val, diag := attr.Expr.Value(childctx)
			if diag.HasErrors() {
				return diag
			}
			val, _ = val.Unmark()
			parentctx.ApplyKeyVal(attr.Name, val, attr.NameRange)
			return hcl.Diagnostics{}
		}

		nme := "for:" + e.StartRange().String()

		child := parentctx.NewChild(nme, e.StartRange())

		diags := hcl.Diagnostics{}

		attrn := NewForCollectionAttribute(e.CollExpr)
		diag := EvaluateAttr(ctx, attrn, child)
		diags = append(diags, diag...)
		delete(parentctx.Map, nme)

		if len(diags) > 0 {
			return diags
		}

		e.CollExpr = &PreCalcExpr{
			Expression: e.CollExpr,
			Val:        child.Map["for_collection"].ToValue(),
		}

		e.ValExpr = &AugmentedForValueExpr{
			Expression: e.ValExpr,
			ForExpr:    e,
			Sudo:       child,
			Ctx:        ctx,
		}

		if e.KeyExpr != nil {
			e.KeyExpr = &AugmentedForValueExpr{
				Expression: e.KeyExpr,
				ForExpr:    e,
				Sudo:       child,
				Ctx:        ctx,
			}
		}

		return EvaluateAttr(ctx, attr, parentctx)

	case *hclsyntax.FunctionCallExpr:

		if len(e.Args) == 0 {
			return EvaluateAttr(ctx, attr, parentctx)
		}

		if _, ok := e.Args[0].(*PreCalcExpr); ok {
			val, diag := attr.Expr.Value(childctx)
			if diag.HasErrors() {
				return diag
			}
			val, _ = val.Unmark()
			parentctx.ApplyKeyVal(attr.Name, val, attr.NameRange)
			return hcl.Diagnostics{}
		}
		// val, _ = val.Unmark()

		child := parentctx.NewChild(FuncKey+":"+e.StartRange().String(), e.StartRange())

		diags := hcl.Diagnostics{}

		newargs := map[hclsyntax.Expression]hclsyntax.Expression{}

		for i, v := range e.Args {
			attrn := NewFuncArgAttribute(i, v)
			diag := EvaluateAttr(ctx, attrn, child)
			diags = append(diags, diag...)
		}

		if len(diags) > 0 {
			return diags
		}

		for k, v := range e.Args {
			argsud := child.Map[fmt.Sprintf("func_arg:%d", k)]
			newargs[v] = &PreCalcExpr{
				Expression: v,
				Val:        argsud.ToValue().Mark(v.Range()),
			}
		}

		for k, v := range e.Args {
			e.Args[k] = newargs[v]
		}

		return EvaluateAttr(ctx, attr, parentctx)

	default:

		val, diag := attr.Expr.Value(childctx)
		if diag.HasErrors() {
			return diag
		}

		// we want to remark this value with the name attirbutes
		val, _ = val.Unmark()

		parentctx.ApplyKeyVal(attr.Name, val, attr.NameRange)
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
	start := 3

	for ((len(retrys) > 0 || len(retryattrs) > 0) && (len(prevRetrys) > len(retrys) || len(prevAttrRetrys) > len(retryattrs))) || start > 0 {

		start--
		newRetrys := []*hclsyntax.Block{}
		newAttrRetrys := []*attr{}
		diags := hcl.Diagnostics{}

		for v := range len(retryattrs) + len(retrys) {
			if v < len(retryattrs) {
				attr := retryattrs[v]
				diagd := EvaluateAttr(ctx, attr.Attri, parentctx)
				if diagd.HasErrors() {
					diags = append(diags, diagd...)
					newAttrRetrys = append(newAttrRetrys, attr)
				}
			} else {
				attr := retrys[v-len(retryattrs)]
				// var save *hclsyntax.Attribute
				// if attr.Type == "gen" {
				// 	save = attr.Body.Attributes["data"]
				// 	delete(attr.Body.Attributes, "data")
				// }
				diagd := NewUnknownBlockEvaluation(ctx, parentctx, attr)
				// if save != nil {
				// 	attr.Body.Attributes["data"] = save
				// }
				if diagd.HasErrors() {
					diags = append(diags, diagd...)
					newRetrys = append(newRetrys, attr)
				}
			}
		}

		if len(diags) < len(lastDiags) {
			start = 3
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

	// strs := []string{block.Type}
	// strs = append(strs, block.Labels...)

	child := parentctx.NewChild(block.Type, block.TypeRange).NewNestedChildBlockLabels(block.Labels, block.LabelRanges)

	for _, attr := range block.Body.Blocks {
		// just init the children so we know that one is missing if it is not completed
		childr := child.NewChild(attr.Type, attr.TypeRange).NewNestedChildBlockLabels(attr.Labels, attr.LabelRanges)
		childr.Meta = &IncomleteBlockMeta{attr}
	}

	userfuncs, _, diag := userfunc.DecodeUserFunctions(block.Body, "func", child.BuildStaticEvalContext)
	if diag.HasErrors() {
		return diag
	}

	child.UserFuncs = userfuncs
	// child.ApplyBlock(block)
	diag = ExtractVariables(ctx, block.Body, child)
	if diag.HasErrors() {
		return diag
	}

	var metad Meta
	blkmeta := &BasicBlockMeta{
		HCL: block,
	}
	metad = blkmeta

	// meta := map[string]cty.Value{
	// 	"label": cty.StringVal(strings.Join(block.Labels, ".")),
	// }

	if block.Type == "gen" {
		um, _ := child.Map["path"].Value.Unmark()

		metad = &GenBlockMeta{
			BasicBlockMeta: *blkmeta,
			RootRelPath:    sanatizeGenPath(um.AsString()),
		}
	}

	child.Meta = metad

	return hcl.Diagnostics{}
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
		"merge":                  ChildMarkMergeFunc,
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

	str = filepath.Base(str)

	if !strings.HasSuffix(str, ".retab") {
		str = str + ".retab"
	}

	// return strings.TrimPrefix(str, ".retab/")

	return str
}

func NewContextualizedFunctionMap(ectx *SudoContext, file string) map[string]function.Function {

	file = sanitizeFileName(file)

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

			unmarked, mrk := args[0].Unmark()

			fle := ectx.Root().Map[FilesKey].Map[file]

			if fle == nil {
				return cty.NilVal, terrors.Errorf("file %s not found", file)
			}

			ok := fle.BlocksOfType(unmarked.AsString())
			if len(ok) == 0 {
				return cty.NilVal, terrors.Errorf("block %s not found", unmarked.AsString())
			}

			err = CheckForAnyIncompletedBlock(ok)
			if err != nil {
				return cty.NilVal, err
			}

			resp := make(map[string]cty.Value, len(ok))
			for _, v := range ok {
				resp[v.ParentKey] = v.ToValue()
			}

			return cty.ObjectVal(resp).WithMarks(mrk), nil
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

			unmarked, mrk := args[0].Unmark()

			// we do not care whether the file is complete or not, just the internal blocks
			fle := ectx.Root().Map[FilesKey].Map[file]

			if fle == nil {
				return cty.NilVal, terrors.Errorf("file %s not found", file)
			}

			ok := fle.BlocksOfType(unmarked.AsString())
			if len(ok) == 0 {
				return cty.NilVal, terrors.Errorf("block %s not found", unmarked.AsString())
			}

			err = CheckForAnyIncompletedBlock(ok)
			if err != nil {
				return cty.NilVal, err
			}

			resp := make([]cty.Value, len(ok))
			for i, v := range ok {
				resp[i] = v.ToValue()
			}

			return cty.TupleVal(resp).WithMarks(mrk), nil
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

			unmarked, mrk := args[0].Unmark()

			fne := sanitizeFileName(unmarked.AsString())

			resp, err := CheckForCompletedBlock(ectx.Map[FilesKey], fne)
			if err != nil {
				return cty.NilVal, err
			}

			return resp.ToValueWithExtraContext().WithMarks(mrk), nil
		},
	})

	return map[string]function.Function{
		"file":       filed,
		"allof":      mapp,
		"alloflist":  list,
		"allofarray": list,
	}
}

func CheckForCompletedBlock(ectx *SudoContext, file string) (*SudoContext, error) {
	resp := ectx.Map[file]

	if resp == nil {
		options := []string{}
		for k := range ectx.Map {
			options = append(options, k)
		}
		return nil, terrors.Errorf("block %q not found: (options: %v)", file, options)
	}

	_, ok := resp.Meta.(*IncomleteBlockMeta)
	if ok {
		return nil, terrors.Errorf("the block %q is not complete", file)
	}
	return resp, nil
}

func CheckForAnyIncompletedBlock(ectx []*SudoContext) error {

	for _, v := range ectx {
		// fmt.Println(v.ParentKey, reflect.TypeOf(v.Meta).String())
		_, okd := v.Meta.(*IncomleteBlockMeta)
		if okd {
			return terrors.Errorf("the block %s is not complete", v.ParentKey)
		}
	}

	return nil
}

func NewDynamicContextualizedFunctionMap(ectx *SudoContext) map[string]function.Function {
	// takes in some negative number and returns the nested parent -x levels
	selfer := function.New(&function.Spec{
		Description: `Returns the parent block of the current block`,
		Params:      []function.Parameter{},
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
