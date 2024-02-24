package lang

import (
	"context"
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/walteh/yaml"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

type SudoContext struct {
	ParentKey string
	Parent    *SudoContext
	Map       map[string]*SudoContext
	// Array     []*SudoContext
	Value     *cty.Value
	isArray   bool
	UserFuncs map[string]function.Function
	Meta      Meta
}

type RemappableSudoContextArray []*SudoContext

func (me *SudoContext) ApplyValue(met cty.Value, r hcl.Range) {
	// val := met.Value()
	// if val.Type().IsObjectType() {
	// 	vm := map[string]Meta{}
	// 	for k, v := range val.AsValueMap() {
	// 		vm[k] = NewSimpleKeyMeta(me, -1, v)
	// 	}
	// } else {
	resp := met.Mark(r)
	me.Value = &resp
	// me.Meta = met
	// }/
}

func (me *SudoContext) ApplyKeyVal(key string, val cty.Value, r hcl.Range) {
	me.NewChild(key).ApplyValue(val, r)
}

// func (me *SudoContext) ApplyValueMap(val map[string]Meta) {
// 	for k, v := range val {
// 		me.ApplyKeyVal(k, v)
// 	}
// }

// func (me *SudoContext) ApplyBlock(block *hclsyntax.Block) {
// 	me.Meta = &BasicBlockMeta{HCL: block}
// }

// func (me *SudoContext) ApplyAttr(attr *hclsyntax.Attribute) {
// 	me.Meta = &AttrMeta{HCL: attr}
// }

func (parent *SudoContext) ApplyBody(ctx context.Context, body *hclsyntax.Body) hcl.Diagnostics {
	return ExtractVariables(ctx, body, parent)
}

type SudoContextArray []*SudoContext
type SudoContextMap map[string]*SudoContext

func (me *SudoContext) ToValue() cty.Value {
	// if me.Value != nil {
	// 	return me.Value
	// }

	if me.Value != nil {
		return *me.Value
	}

	if me.isArray || strings.HasPrefix(me.ParentKey, FuncKey) {
		return me.List().ToValue()
	}

	return SudoContextMap(me.Map).ToValue()
}

func (me SudoContextArray) ToValue() cty.Value {
	vals := make([]cty.Value, len(me))
	for i, v := range me {
		if v.Meta == nil {
			vals[i] = v.ToValue()
		} else {

			vals[i] = v.ToValue().Mark(v.Meta.Range())
		}
	}
	return cty.TupleVal(vals)
}

func (me SudoContextMap) ToValue() cty.Value {
	obj := make(map[string]cty.Value, len(me))
	for k, v := range me {
		if v.Meta == nil {
			obj[k] = v.ToValue()
		} else {
			obj[k] = v.ToValue().Mark(v.Meta.Range())
		}
	}
	return cty.ObjectVal(obj)
}

func (me *SudoContext) List() SudoContextArray {
	if me.isArray || strings.HasPrefix(me.ParentKey, FuncKey) {
		return sortMappedIntegerKeys(me.Map)
	}

	return SudoContextMap(me.Map).List()
}

func (me *SudoContext) BuildStaticEvalVars() map[string]cty.Value {
	wrk := map[string]cty.Value{}
	for k, v := range me.Map {
		wrk[k] = v.ToValue()
	}

	return wrk
}

func (me SudoContextMap) List() SudoContextArray {
	ctxs := make([]*SudoContext, 0, len(me))
	for _, v := range me {
		ctxs = append(ctxs, v)
	}

	slices.SortFunc(ctxs, func(a, b *SudoContext) int {
		if a.Meta == nil || b.Meta == nil {
			return 0
		}
		if a.Meta.Range().Start.Line == b.Meta.Range().Start.Line {
			return a.Meta.Range().Start.Column - b.Meta.Range().Start.Column
		}
		return a.Meta.Range().Start.Line - b.Meta.Range().Start.Line
	})
	return ctxs
}

func sortMappedIntegerKeys[K any](m map[string]K) []K {
	type sorter struct {
		Key   int
		Value K
	}

	var keys []sorter
	for k := range m {
		splt := strings.Split(k, ":")
		slices.Reverse(splt)
		i, _ := strconv.Atoi(splt[0])
		keys = append(keys, sorter{
			Key:   i,
			Value: m[k],
		})
	}

	slices.SortFunc(keys, func(a, b sorter) int {
		return a.Key - b.Key
	})

	var values []K
	for _, k := range keys {
		values = append(values, k.Value)
	}
	return values
}

func (me *SudoContext) BuildStaticVarsList() []cty.Value {

	ctxs := me.List()
	vals := make([]cty.Value, len(ctxs))
	for i, v := range ctxs {
		vals[i] = v.ToValue()
	}

	return vals

}

func (me *SudoContext) Functions() map[string]function.Function {
	fn := NewFunctionMap()

	if me.UserFuncs != nil {
		for k, v := range me.UserFuncs {
			fn[k] = v
		}
	}

	return fn
}

func (me *SudoContext) NewNestedChildBlockLabels(key []string, ranges []hcl.Range) *SudoContext {
	wrk := me
	for i, v := range key {
		wrk = wrk.NewChild(v)
		wrk.Meta = &BlockLabelMeta{HCL: ranges[i]}
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

	internalParentz, _ := internalParent[file].Unmark()

	internalParent = internalParentz.AsValueMap()

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

	for k, v := range NewContextualizedFunctionMap(wc.Root(), file) {
		wrk.Functions[k] = v
	}

	for k, v := range NewDynamicContextualizedFunctionMap(wc) {
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

func ToYAML(ctx *SudoContext) (yaml.MapSlice, error) {
	cnt := yaml.MapSlice{}

	slc, err := ctx.ToYAML()
	if err != nil {
		return nil, err
	}

	if x, ok := slc.(yaml.MapSlice); ok {
		cnt = append(cnt, x...)
	}
	if x, ok := slc.(yaml.MapItem); ok {
		cnt = append(cnt, x)
	}

	return cnt, nil
}
func (me *SudoContext) ToYAML() (any, error) {
	if me.Value != nil {
		// if val, ok := me.Value.(*AttrMeta); ok {
		enc, err := noMetaJsonEncode(*me.Value)
		if err != nil {
			return nil, err
		}
		return enc, nil
		// }
	}

	if me.isArray {
		return me.List().ToYAML()
	}

	return SudoContextMap(me.Map).ToYAML()
}

func (me SudoContextArray) ToYAML() (any, error) {
	out := make([]any, len(me))
	for i, v := range me {
		enc, err := v.ToYAML()
		if err != nil {
			return nil, err
		}
		out[i] = enc
	}
	return out, nil
}

func (me SudoContextMap) ToYAML() (any, error) {
	wrk := make(yaml.MapSlice, 0)

	list := me.List()
	for _, g := range list {
		if g.ParentKey == MetaKey || strings.Contains(g.ParentKey, FuncKey) {
			continue
		}
		yml, err := g.ToYAML()
		if err != nil {
			return nil, err
		}
		wrk = append(wrk, yaml.MapItem{Key: g.ParentKey, Value: yml})
	}

	return wrk, nil
}

func (me *SudoContext) BlocksOfType(name string) SudoContextArray {
	blks := []*SudoContext{}
	for k, v := range me.Map {
		if k == name {
			for _, d := range v.Map {
				for _, ok := d.Meta.(*BlockLabelMeta); ok; _, ok = d.Meta.(*BlockLabelMeta) {
					d = d.List()[0]
				}
				if bm, ok := d.Meta.(BlockMeta); ok {
					if bm.Block().Type == name {
						blks = append(blks, d)
					}
				}
			}
		}
	}
	return blks
}

func (me *SudoContext) GetAllFileLevelBlocksOfType(name string) SudoContextArray {
	files := me.Map[FilesKey].Map
	out := []*SudoContext{}
	for _, v := range files {
		for _, blk := range v.BlocksOfType(name) {
			out = append(out, blk)
		}
	}
	return out
}

// func valRange(z cty.Value) (cty.Value, hcl.Range) {
// 	me, _ := z.Unmark()
// 	rnge := me.EncapsulatedValue().(hcl.Range)

//		return me, rnge
//	}
func valRange(me cty.Value) (cty.Value, hcl.Range) {
	var rng hcl.Range
	me, r := me.Unmark()
	for z := range r {
		if intre, ok := z.(hcl.Range); ok {
			rng = intre
		}
	}
	// okay := any(r)

	// for _, v := range r {
	// 	fmt.Println(v)
	// }

	return me, rng
}

func UnmarkToSortedArray(me cty.Value) (any, error) {
	me, _ = valRange(me)

	if me.Type().IsObjectType() {
		type Sorter struct {
			Range hcl.Range
			Key   string
			Value cty.Value
		}

		objs := me.AsValueMap()
		out := make([]Sorter, 0, len(objs))
		for k, v := range objs {
			v, rng := valRange(v)
			out = append(out, Sorter{Range: rng, Key: k, Value: v})
		}
		slices.SortFunc(out, func(a, b Sorter) int {
			if a.Range.Start.Line == b.Range.Start.Line {
				return a.Range.Start.Byte - b.Range.Start.Byte
			}
			return a.Range.Start.Line - b.Range.Start.Line
		})
		wrk := make(yaml.MapSlice, 0, len(out))
		for _, v := range out {
			res, err := UnmarkToSortedArray(v.Value)
			if err != nil {
				return nil, err
			}
			wrk = append(wrk, yaml.MapItem{Key: v.Key, Value: res})
		}
		return wrk, nil
	}

	if me.Type().IsTupleType() {
		type Sorter struct {
			Range hcl.Range
			Value cty.Value
		}

		objs := me.AsValueSlice()
		out := make([]Sorter, 0, len(objs))
		for _, v := range objs {
			v, rng := valRange(v)
			out = append(out, Sorter{Range: rng, Value: v})
		}
		slices.SortFunc(out, func(a, b Sorter) int {
			if a.Range.Start.Line == b.Range.Start.Line {
				return a.Range.Start.Byte - b.Range.Start.Byte
			}
			return a.Range.Start.Line - b.Range.Start.Line
		})
		wrk := make([]any, 0, len(out))
		for _, v := range out {
			res, err := UnmarkToSortedArray(v.Value)
			if err != nil {
				return nil, err
			}
			wrk = append(wrk, res)
		}
		return wrk, nil
	}

	return noMetaJsonEncode(me)

}
