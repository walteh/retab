package lang

import (
	"context"
	"strconv"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func NewFuncArgAttribute(index int, expr hclsyntax.Expression) *hclsyntax.Attribute {
	return &hclsyntax.Attribute{
		Name:      "func_arg:" + strconv.Itoa(index),
		Expr:      expr,
		NameRange: expr.Range(),
		SrcRange:  expr.Range(),
	}
}

func NewArrayItemAttribute(index int, expr hclsyntax.Expression) *hclsyntax.Attribute {
	return &hclsyntax.Attribute{
		Name:      "array_item:" + strconv.Itoa(index),
		Expr:      expr,
		NameRange: expr.Range(),
		SrcRange:  expr.Range(),
	}
}

func NewObjectItemAttribute(key string, nameRange hcl.Range, expr hclsyntax.Expression) *hclsyntax.Attribute {
	return &hclsyntax.Attribute{
		Name:      key,
		Expr:      expr,
		NameRange: nameRange,
		SrcRange:  expr.Range(),
	}
}

func NewForCollectionAttribute(expr hclsyntax.Expression) *hclsyntax.Attribute {
	return &hclsyntax.Attribute{
		Name:      "for_collection",
		Expr:      expr,
		NameRange: expr.Range(),
		SrcRange:  expr.Range(),
	}
}

type WrappedExpression struct {
	hclsyntax.Node
	Expr hclsyntax.Expression
	Sudo *SudoContext
}

// Range implements hclsyntax.Expression.
func (me *WrappedExpression) Range() hcl.Range {
	return me.Expr.Range()
}

// StartRange implements hclsyntax.Expression.
func (me *WrappedExpression) StartRange() hcl.Range {
	return me.Expr.StartRange()
}

// Value implements hclsyntax.Expression.
func (me *WrappedExpression) Value(ectx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	val, diag := me.Expr.Value(ectx)
	if diag.HasErrors() {
		return cty.DynamicVal, diag
	}
	return val.Mark(me.Sudo.Meta.Range()), diag
}

// Variables implements hclsyntax.Expression.
func (me *WrappedExpression) Variables() []hcl.Traversal {
	return me.Expr.Variables()
}

var _ hclsyntax.Expression = (*WrappedExpression)(nil)

type PreCalcExpr struct {
	hclsyntax.Expression
	Val cty.Value
}

// Range implements hclsyntax.Expression.
func (me *PreCalcExpr) Range() hcl.Range {
	return me.Expression.Range()
}

// StartRange implements hclsyntax.Expression.
func (me *PreCalcExpr) StartRange() hcl.Range {
	return me.Expression.StartRange()
}

// Value implements hclsyntax.Expression.
func (me *PreCalcExpr) Value(ectx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	return me.Val, hcl.Diagnostics{}
}

// Variables implements hclsyntax.Expression.
func (me *PreCalcExpr) Variables() []hcl.Traversal {
	return me.Expression.Variables()
}

type AugmentedForValueExpr struct {
	hclsyntax.Expression
	Sudo    *SudoContext
	ForExpr *hclsyntax.ForExpr
	Ctx     context.Context
}

// Range implements hclsyntax.Expression.
func (me *AugmentedForValueExpr) Range() hcl.Range {
	return me.Expression.Range()
}

// StartRange implements hclsyntax.Expression.
func (me *AugmentedForValueExpr) StartRange() hcl.Range {
	return me.Expression.StartRange()
}

// Value implements hclsyntax.Expression.
func (me *AugmentedForValueExpr) Value(ectx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {

	child := me.Sudo.NewChild("tmp", me.StartRange())

	// no need for the parent to have a reference to the child
	// delete(me.Sudo.Map, "tmp")
	// child.Parent = nil
	var mrk hcl.Range

	srtr := NewSorter("", ectx.Variables[me.ForExpr.ValVar])
	if mrk.Empty() {
		mrk = srtr.Range[0]
	}

	if me.ForExpr.KeyVar != "" {
		// mrk = ectx.Variables[me.ForExpr.KeyVar].Marks()
		child.TmpFileLevelVars[me.ForExpr.KeyVar] = ectx.Variables[me.ForExpr.KeyVar]
	}

	child.TmpFileLevelVars[me.ForExpr.ValVar] = ectx.Variables[me.ForExpr.ValVar]

	// defer func() {
	// 	delete(child.TmpFileLevelVars, me.ForExpr.ValVar)
	// 	delete(child.TmpFileLevelVars, me.ForExpr.KeyVar)
	// }()

	diags := EvaluateAttr(me.Ctx, NewForCollectionAttribute(me.Expression), child)

	if diags.HasErrors() {
		return cty.DynamicVal, diags
	}

	delete(me.Sudo.Map, "tmp")

	vals := child.Map["for_collection"].ToValue()

	// delete(child.Map, "for_collection")

	return vals.Mark(mrk), hcl.Diagnostics{}

	// aug, diags := me.Expression.Value(ectx)
	// return aug.WithMarks(mrk), diags
}

// Variables implements hclsyntax.Expression.
func (me *AugmentedForValueExpr) Variables() []hcl.Traversal {
	return me.Expression.Variables()
}
