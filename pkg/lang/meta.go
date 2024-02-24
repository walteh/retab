package lang

import (
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

func NewObjectItemAttribute(key string, expr hclsyntax.Expression) *hclsyntax.Attribute {
	return &hclsyntax.Attribute{
		Name:      key,
		Expr:      expr,
		NameRange: expr.Range(),
		SrcRange:  expr.Range(),
	}
}

// type WrappedExpression struct {
// 	hclsyntax.Node
// 	Expr hclsyntax.Expression
// 	Sudo *SudoContext
// }

// Range implements hclsyntax.Expression.
// func (me *WrappedExpression) Range() hcl.Range {
// 	return me.Expr.Range()
// }

// // StartRange implements hclsyntax.Expression.
// func (me *WrappedExpression) StartRange() hcl.Range {
// 	return me.Expr.StartRange()
// }

// // Value implements hclsyntax.Expression.
// func (me *WrappedExpression) Value(_ *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
// 	return me.Expr.Value(me.Sudo.BuildStaticEvalContextWithFileData(me.Range().Filename))
// }

// // Variables implements hclsyntax.Expression.
// func (me *WrappedExpression) Variables() []hcl.Traversal {
// 	return me.Expr.Variables()
// }

// var _ hclsyntax.Expression = (*WrappedExpression)(nil)
