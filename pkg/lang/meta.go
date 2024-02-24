package lang

import (
	"context"
	"strconv"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type Meta interface {
	// Range() hcl.Range
	// Value() cty.Value
	Range() hcl.Range
}

var (
	_ Meta      = (*BasicBlockMeta)(nil)
	_ Meta      = (*AttrMeta)(nil)
	_ Meta      = (*GenBlockMeta)(nil)
	_ Meta      = (*SimpleNameMeta)(nil)
	_ BlockMeta = (*BasicBlockMeta)(nil)
	_ BlockMeta = (*GenBlockMeta)(nil)
)

type BlockMeta interface {
	Meta
	Block() *hclsyntax.Block
}

type BasicBlockMeta struct {
	HCL *hclsyntax.Block
}

func (me *BasicBlockMeta) Block() *hclsyntax.Block {
	return me.HCL
}

func (me *BasicBlockMeta) Range() hcl.Range {
	return me.HCL.TypeRange
}

type AttrMeta struct {
	HCL   *hclsyntax.Attribute
	Value cty.Value
}

func (me *AttrMeta) Range() hcl.Range {
	return me.HCL.SrcRange
}

type GenBlockMeta struct {
	HCL         *hclsyntax.Block
	RootRelPath string
}

type SimpleNameMeta struct {
	NameRange hcl.Range
}

func NewSimpleNameMeta(parent hcl.Range) *SimpleNameMeta {
	return &SimpleNameMeta{parent}
}

func (me *SimpleNameMeta) Range() hcl.Range {
	return me.NameRange
}

func (me *GenBlockMeta) Range() hcl.Range {
	return me.HCL.TypeRange
}

func (me *GenBlockMeta) Block() *hclsyntax.Block {
	return me.HCL
}

type BlockLabelMeta struct {
	HCL hcl.Range
}

func (me *BlockLabelMeta) Range() hcl.Range {
	return me.HCL
}

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

	var mrk hcl.Range

	srtr := valRange("", ectx.Variables[me.ForExpr.ValVar])
	// child.ApplyKeyVal(me.ForExpr.ValVar, v, mrkd)
	if mrk.Empty() {
		mrk = srtr.Range[0]
	}

	// if me.ForExpr.KeyVar != "" {
	// 	// v, mrkd := valRange(ectx.Variables[me.ForExpr.KeyVar])
	// 	// child.ApplyKeyVal(me.ForExpr.KeyVar, ectx.Variables[me.ForExpr.KeyVar], mrk)
	// 	// mrk = mrkd
	// }

	// suppEctx := &hcl.EvalContext{
	// 	Variables: make(map[string]cty.Value),
	// 	Functions: make(map[string]function.Function),
	// }

	// var mrk cty.ValueMarks

	if me.ForExpr.KeyVar != "" {
		// mrk = ectx.Variables[me.ForExpr.KeyVar].Marks()
		child.TmpFileLevelVars[me.ForExpr.KeyVar] = ectx.Variables[me.ForExpr.KeyVar]
	}

	// if len(mrk) == 0 {
	// 	mrk = ectx.Variables[me.ForExpr.ValVar].Marks()
	// }

	child.TmpFileLevelVars[me.ForExpr.ValVar] = ectx.Variables[me.ForExpr.ValVar]

	diags := EvaluateAttr(me.Ctx, NewForCollectionAttribute(me.Expression), child)
	delete(me.Sudo.Map, "tmp")

	if diags.HasErrors() {
		return cty.DynamicVal, diags
	}

	vals := child.Map["for_collection"].ToValue()

	return vals.Mark(mrk), hcl.Diagnostics{}

	// aug, diags := me.Expression.Value(ectx)
	// return aug.WithMarks(mrk), diags
}

// Variables implements hclsyntax.Expression.
func (me *AugmentedForValueExpr) Variables() []hcl.Traversal {
	return me.Expression.Variables()
}