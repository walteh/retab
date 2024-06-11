package lang

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

const UserFuncBlockType = "func"

func NewContextFromFiles(ctx context.Context, fle map[string][]byte, env map[string]string) (*hcl.File, *SudoContext, *BodyBuilder, hcl.Diagnostics, error) {

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

	for k, v := range env {
		mectx.TmpFileLevelVars[k] = cty.StringVal(v)
	}

	diags := mectx.ApplyBody(ctx, root.NewRoot())

	return nil, mectx, root, diags, nil

}

func NewContextFromFile(ctx context.Context, fle []byte, name string) (*hcl.File, *SudoContext, *BodyBuilder, hcl.Diagnostics, error) {
	return NewContextFromFiles(ctx, map[string][]byte{sanitizeFileName(name): fle}, map[string]string{})
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

		child, err := parentctx.NewNonBlockChild(attr.Name, attr.NameRange)
		if err != nil {
			return hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "block already exists with this name",
					Detail:   err.Error(),
					Subject:  &attr.NameRange,
				},
			}
		}

		diags := hcl.Diagnostics{}
		for _, v := range e.Items {
			key, diag := v.KeyExpr.Value(childctx)
			if diag.HasErrors() {
				diags = append(diags, diag...)
				continue
			}

			key, diag = unmarkCheckingForIncompleteBlock(key)
			if diag.HasErrors() {
				diags = append(diags, diag...)
				continue
			}

			attrn := NewObjectItemAttribute(key.AsString(), v.KeyExpr.Range(), v.ValueExpr)
			attrn.NameRange = v.ValueExpr.Range()

			diag = EvaluateAttr(ctx, attrn, child, extra...)
			if diag.HasErrors() {
				diags = append(diags, diag...)
			}

			if len(diags) == 0 {

				child.Map[key.AsString()].Meta = &SimpleNameMeta{v.KeyExpr.Range()}
			}
		}

		return diags
	case *hclsyntax.TupleConsExpr:

		child, err := parentctx.NewNonBlockChild(attr.Name, attr.NameRange)
		if err != nil {
			return hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "block already exists with this name",
					Detail:   err.Error(),
					Subject:  &attr.NameRange,
				},
			}
		}

		child.isArray = true

		diags := hcl.Diagnostics{}

		for i, v := range e.Exprs {
			attrn := NewArrayItemAttribute(i, v)
			attrn.NameRange = v.Range()
			diag := EvaluateAttr(ctx, attrn, child, extra...)
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
	case *hclsyntax.ForExpr:

		if pce, ok := e.CollExpr.(*PreCalcExpr); ok {
			// if vce, ok := e.ValExpr.(*AugmentedForValueExpr); ok {

			val, diag := attr.Expr.Value(childctx)
			if diag.HasErrors() {
				return diag
			}
			val, diags := unmarkCheckingForIncompleteBlock(val)
			if diags.HasErrors() {
				return diags
			}

			// we have to reset or subsequent calls will break
			// this resolves issue #33
			e.CollExpr = pce.Expression

			return parentctx.ApplyKeyVal(attr.Name, val, attr.NameRange)
		}

		nme := "for:" + e.StartRange().String()

		// no need to check if this is already a block, it is not possible
		child := parentctx.NewChild(nme, e.CollExpr.Range())

		diags := hcl.Diagnostics{}

		attrn := NewForCollectionAttribute(e.CollExpr)
		diag := EvaluateAttr(ctx, attrn, child, extra...)
		diags = append(diags, diag...)

		if len(diags) > 0 {
			return diags
		}

		// to prevent the child from being used in the parent
		// we only do it after the diags check - that way evaluation is not duplicated
		delete(parentctx.Map, nme)

		vald := child.Map["for_collection"].ToValue()

		e.CollExpr = &PreCalcExpr{
			Expression: e.CollExpr,
			Val:        vald,
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

		return EvaluateAttr(ctx, attr, parentctx, extra...)

	case *hclsyntax.FunctionCallExpr:

		if len(e.Args) == 0 {
			return EvaluateAttr(ctx, attr, parentctx, extra...)
		}

		if _, ok := e.Args[0].(*PreCalcExpr); ok {
			val, diag := attr.Expr.Value(childctx)
			if diag.HasErrors() {
				return diag
			}
			val, diags := unmarkCheckingForIncompleteBlock(val)
			if diags.HasErrors() {
				return diags
			}
			return parentctx.ApplyKeyVal(attr.Name, val, attr.NameRange)
		}
		// val, _ = val.Unmark()

		child := parentctx.NewChild(FuncKey+":"+e.StartRange().String(), e.StartRange())

		diags := hcl.Diagnostics{}

		newargs := map[hclsyntax.Expression]hclsyntax.Expression{}

		for i, v := range e.Args {
			attrn := NewFuncArgAttribute(i, v)
			diag := EvaluateAttr(ctx, attrn, child, extra...)
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

		return EvaluateAttr(ctx, attr, parentctx, extra...)

	default:

		val, diag := attr.Expr.Value(childctx)
		if diag.HasErrors() {
			return diag
		}

		// we want to remark this value with the name attirbutes
		val, diags := unmarkCheckingForIncompleteBlock(val)
		if diags.HasErrors() {
			return diags
		}

		return parentctx.ApplyKeyVal(attr.Name, val, attr.NameRange)
	}

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
				diagd := NewUnknownBlockEvaluation(ctx, parentctx, attr)
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

	if block.Type == UserFuncBlockType {
		// since we go back to the parent, we don't need to process the functions if they are already there
		// if len(parentctx.Parent.UserFuncs) == 0 {
		// 	userfuncs, diags := ExtractUserFuncs(ctx, block.Body, parentctx.Parent)
		// 	if diags.HasErrors() {
		// 		return diags
		// 	}

		// 	parentctx.Parent.UserFuncs = userfuncs

		// 	return diags
		// }

		// skip normal processing of extra user functions
		return hcl.Diagnostics{}
	}

	child, err := parentctx.NewBlockChild(block)
	if err != nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "failed to create block",
				Detail:   err.Error(),
				Subject:  &block.TypeRange,
			},
		}
	}

	userfuncs, _, diag := decodeUserFunctions(ctx, block.Body, UserFuncBlockType, child.BuildStaticEvalContext, child)
	if diag.HasErrors() {
		return diag
	}

	child.UserFuncs = userfuncs
	diag = ExtractVariables(ctx, block.Body, child)
	if diag.HasErrors() {
		return diag
	}

	var metad Meta
	blkmeta := &BasicBlockMeta{
		HCL: block,
	}
	metad = blkmeta

	if block.Type == "gen" {
		if child.Map["path"] == nil || child.Map["path"].Value == nil {
			return hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "missing attribute",
					Detail:   "a gen block must have a path attribute",
					Subject:  &block.TypeRange,
				},
			}
		}
		um, diags := unmarkCheckingForIncompleteBlock(*child.Map["path"].Value)
		if diags.HasErrors() {
			return diags
		}

		metad = &GenBlockMeta{
			BasicBlockMeta: *blkmeta,
			RootRelPath:    sanatizeGenPath(um.AsString()),
		}
	}

	child.Meta = metad

	return hcl.Diagnostics{}
}

func unmarkCheckingForIncompleteBlock(val cty.Value) (cty.Value, hcl.Diagnostics) {

	markstokeep := cty.ValueMarks{}

	mrks := val.Marks()
	if len(mrks) > 0 {
		for v := range mrks {
			if _, ok := v.(isIncompleteBlock); ok {
				return cty.NilVal, hcl.Diagnostics{
					{
						Severity: hcl.DiagError,
						Summary:  "incomplete block",
						Detail:   "this block is incomplete",
					},
				}
			}
			// if _, ok := v.(*ignoreFromYaml); ok {
			// 	markstokeep[v] = struct{}{}
			// }
			if z, ok := v.(*makePathRelative); ok {
				markstokeep[z] = struct{}{}
				// val = cty.StringVal("CAN_BE_" + val.AsString() + "RELATIVE_" z)
			}
		}
	}

	unmrk, _ := val.Unmark()

	return unmrk.WithMarks(markstokeep), hcl.Diagnostics{}
}
