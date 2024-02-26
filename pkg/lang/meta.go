package lang

import (
	"context"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type Meta interface {
	// Range() hcl.Range
	// Value() cty.Value
	Range() hcl.Range
	Variables() map[string]cty.Value
}

var (
	_ Meta      = (*BasicBlockMeta)(nil)
	_ Meta      = (*AttrMeta)(nil)
	_ Meta      = (*IncomleteBlockMeta)(nil)
	_ BlockMeta = (*IncomleteBlockMeta)(nil)
	_ Meta      = (*GenBlockMeta)(nil)
	_ Meta      = (*SimpleNameMeta)(nil)
	_ BlockMeta = (*BasicBlockMeta)(nil)
	_ BlockMeta = (*GenBlockMeta)(nil)
	_ Meta      = (*BlockLabelMeta)(nil)
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

type ignoreFromYaml struct{}

func (me *BasicBlockMeta) Variables() map[string]cty.Value {
	vals := make(map[string]cty.Value)

	vals["label"] = cty.StringVal(strings.Join(me.HCL.Labels, ".")).Mark(&ignoreFromYaml{})
	vals["type"] = cty.StringVal(me.HCL.Type).Mark(&ignoreFromYaml{})

	return vals
}

type IncomleteBlockMeta struct {
	HCL *hclsyntax.Block
}

func (me *IncomleteBlockMeta) Block() *hclsyntax.Block {
	return me.HCL
}

func (me *IncomleteBlockMeta) Range() hcl.Range {
	return me.HCL.TypeRange
}

func (me *IncomleteBlockMeta) Variables() map[string]cty.Value {
	return map[string]cty.Value{}
}

type AttrMeta struct {
	HCL   *hclsyntax.Attribute
	Value cty.Value
}

// Variables implements Meta.
func (*AttrMeta) Variables() map[string]cty.Value {
	return map[string]cty.Value{}
}

func (me *AttrMeta) Range() hcl.Range {
	return me.HCL.SrcRange
}

type GenBlockMeta struct {
	BasicBlockMeta
	RootRelPath string
}

type SimpleNameMeta struct {
	NameRange hcl.Range
}

// Variables implements Meta.
func (*SimpleNameMeta) Variables() map[string]cty.Value {
	return map[string]cty.Value{}
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

func (me *GenBlockMeta) Variables() map[string]cty.Value {
	vals := me.BasicBlockMeta.Variables()

	vals["resolved_output"] = cty.StringVal(me.RootRelPath).Mark(&ignoreFromYaml{})
	vals["source"] = cty.StringVal(me.HCL.TypeRange.Filename).Mark(&ignoreFromYaml{})

	return vals
}

type BlockLabelMeta struct {
	HCL hcl.Range
}

func (me *BlockLabelMeta) Range() hcl.Range {
	return me.HCL
}

func (me *BlockLabelMeta) Variables() map[string]cty.Value {
	return map[string]cty.Value{}

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

	// no need for the parent to have a reference to the child
	// delete(me.Sudo.Map, "tmp")
	// child.Parent = nil
	var mrk hcl.Range

	srtr := valRange("", ectx.Variables[me.ForExpr.ValVar])
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
