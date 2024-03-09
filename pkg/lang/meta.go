package lang

import (
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type Meta interface {
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
	Enhance(v cty.Value, ok []any) (cty.Value, error)
}

type BasicBlockMeta struct {
	HCL *hclsyntax.Block
}

// Enhance implements BlockMeta.
func (*BasicBlockMeta) Enhance(v cty.Value, ok []any) (cty.Value, error) {
	return v, nil
}

func (me *BasicBlockMeta) Block() *hclsyntax.Block {
	return me.HCL
}

func (me *BasicBlockMeta) Range() hcl.Range {
	return me.HCL.TypeRange
}

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

func (me *IncomleteBlockMeta) Enhance(v cty.Value, ok []any) (cty.Value, error) {
	panic("can't enhance incomplete block")
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

// Enhance implements SortEnhancer.
func (me *GenBlockMeta) Enhance(v cty.Value, ok []any) (cty.Value, error) {
	for _, v := range ok {
		switch t := v.(type) {
		case *makePathRelative:
			res, err := filepath.Rel(me.RootRelPath, t.to)
			if err != nil {
				return cty.DynamicVal, err
			}
			return cty.StringVal(res).Mark(&ignoreFromYaml{}), nil
		}
	}

	return v, nil
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

	vals["resolved_output"] = cty.StringVal(me.RootRelPath).Mark(&ignoreFromYaml{}).Mark(&makePathRelative{me.RootRelPath})
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
