package lang

import (
	"context"
	"fmt"
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
	ParentKey        string
	Parent           *SudoContext
	Map              map[string]*SudoContext
	Value            *cty.Value
	isArray          bool
	UserFuncs        map[string]function.Function
	Meta             Meta
	TmpFileLevelVars map[string]cty.Value
}

type RemappableSudoContextArray []*SudoContext

func (me *SudoContext) ApplyValue(met cty.Value) {
	me.Value = &met
}

func (me *SudoContext) ApplyKeyVal(key string, val cty.Value, r hcl.Range) {
	me.NewChild(key, r).ApplyValue(val)
}

func (parent *SudoContext) ApplyBody(ctx context.Context, body *hclsyntax.Body) hcl.Diagnostics {
	return ExtractVariables(ctx, body, parent)
}

func (me *SudoContext) ToValueWithExtraContext() cty.Value {

	var val cty.Value

	if me.Value != nil {
		val = *me.Value
	} else {
		if me.isArray || strings.HasPrefix(me.ParentKey, FuncKey) {
			lst := me.List()
			vals := make([]cty.Value, len(lst))
			for i, v := range lst {
				vals[i] = v.ToValueWithExtraContext()
			}
			val = cty.TupleVal(vals)
		} else {
			obj := make(map[string]cty.Value, len(me.Map))
			for k, v := range me.Map {
				obj[k] = v.ToValueWithExtraContext()
			}
			for k, v := range me.Meta.Variables() {
				obj[k] = v
			}
			val = cty.ObjectVal(obj)
		}
	}

	return val.Mark(me.Meta.Range())
}

func (me *SudoContext) ToValue() cty.Value {

	var val cty.Value

	if me.Value != nil {
		val = *me.Value
	} else {
		if me.isArray || strings.HasPrefix(me.ParentKey, FuncKey) {
			lst := me.List()
			vals := make([]cty.Value, len(lst))
			for i, v := range lst {
				vals[i] = v.ToValue()
			}
			val = cty.TupleVal(vals)
		} else {
			obj := make(map[string]cty.Value, len(me.Map))
			for k, v := range me.Map {
				obj[k] = v.ToValue()
			}
			val = cty.ObjectVal(obj)
		}
	}

	return val.Mark(me.Meta.Range())
}

func (me *SudoContext) List() []*SudoContext {
	if me.isArray || strings.HasPrefix(me.ParentKey, FuncKey) {
		return sortMappedIntegerKeys(me.Map)
	}

	ctxs := make([]*SudoContext, 0, len(me.Map))
	for _, v := range me.Map {
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

func (me *SudoContext) BuildStaticEvalVars() map[string]cty.Value {
	obj := make(map[string]cty.Value, len(me.Map))
	for k, v := range me.Map {
		obj[k] = v.ToValueWithExtraContext()
	}

	return obj
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
		vals[i] = v.ToValueWithExtraContext()
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
		wrk = wrk.NewChild(v, ranges[i])
	}
	return wrk
}

func (me *SudoContext) NewChild(key string, rnge hcl.Range) *SudoContext {

	if me.Map[key] != nil {
		me.Map[key].Meta = &SimpleNameMeta{rnge}
		return me.Map[key]
	}

	build := &SudoContext{
		ParentKey:        key,
		Parent:           me,
		Map:              make(map[string]*SudoContext),
		UserFuncs:        make(map[string]function.Function),
		Meta:             &SimpleNameMeta{rnge},
		TmpFileLevelVars: make(map[string]cty.Value),
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

func (wc *SudoContext) GetAllTemporaryFileLevelVars() map[string]cty.Value {
	mine := map[string]cty.Value{}
	if wc.Parent != nil {
		for k, v := range wc.Parent.GetAllTemporaryFileLevelVars() {
			mine[k] = v
		}
	}
	for k, v := range wc.TmpFileLevelVars {
		mine[k] = v
	}
	return mine
}

func (wc *SudoContext) BuildStaticEvalContextWithFileData(file string) *hcl.EvalContext {

	wrk := wc.Root().Map[FilesKey].Map[file].BuildStaticEvalContext()

	for k, v := range wc.Functions() {
		wrk.Functions[k] = v
	}

	for k, v := range NewContextualizedFunctionMap(wc.Root(), file) {
		wrk.Functions[k] = v
	}

	for k, v := range NewDynamicContextualizedFunctionMap(wc) {
		wrk.Functions[k] = v
	}

	for k, v := range wc.GetAllTemporaryFileLevelVars() {
		wrk.Variables[k] = v
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
		lst := me.List()
		out := make([]any, len(lst))
		for i, v := range lst {
			enc, err := v.ToYAML()
			if err != nil {
				return nil, err
			}
			out[i] = enc
		}
		return out, nil
	} else {
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
}

func (me *SudoContext) BlocksOfType(name string) []*SudoContext {
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

func (me *SudoContext) GetAllFileLevelBlocksOfType(name string) []*SudoContext {
	files := me.Map[FilesKey].Map
	out := []*SudoContext{}
	for _, v := range files {
		out = append(out, v.BlocksOfType(name)...)
	}
	return out
}

type Sorter struct {
	Range []hcl.Range
	Key   string
	Value cty.Value
}

func (me *Sorter) Array() []hcl.Range {
	return me.Range
}

func valRange(key string, me cty.Value) *Sorter {

	me, r := me.Unmark()

	if len(r) == 0 {
		panic(fmt.Sprintf("no range found for %s", me.GoString()))
	}

	ranges := make([]hcl.Range, 0, len(r))
	for z := range r {
		if intre, ok := z.(hcl.Range); ok {
			ranges = append(ranges, intre)
		}
	}

	slices.SortFunc(ranges, func(a, b hcl.Range) int {
		if a.Start.Line == b.Start.Line {
			return a.Start.Byte - b.Start.Byte
		}
		return a.Start.Line - b.Start.Line
	})

	return &Sorter{Range: ranges, Key: key, Value: me}
}

func sortem(val cty.Value) []*Sorter {

	out := make([]*Sorter, 0)
	if val.Type().IsObjectType() {
		objs := val.AsValueMap()
		for k, v := range objs {
			rng := valRange(k, v)
			out = append(out, rng)
		}
	} else if val.Type().IsTupleType() {
		objs := val.AsValueSlice()
		for _, v := range objs {
			rng := valRange("", v)
			out = append(out, rng)
		}
	}

	slices.SortFunc(out, func(a, b *Sorter) int {
		for i, x := range a.Range {
			if i >= len(b.Range) {
				return 1
			}
			y := b.Range[i]
			if x.Start.Line == y.Start.Line {
				if x.Start.Column == y.Start.Column {
					continue
				}
				return x.Start.Column - y.Start.Column
			}
			return x.Start.Line - y.Start.Line
		}
		return 0
	})

	return out
}

func UnmarkToSortedArray(me cty.Value) (any, error) {
	me, _ = me.Unmark()

	out := sortem(me)

	if me.Type().IsObjectType() {

		wrk := make(yaml.MapSlice, 0, len(out))
		for _, v := range out {
			if v.Key == MetaKey || strings.Contains(v.Key, FuncKey) {
				continue
			}
			res, err := UnmarkToSortedArray(v.Value)
			if err != nil {
				return nil, err
			}
			wrk = append(wrk, yaml.MapItem{Key: v.Key, Value: res})
		}
		return wrk, nil
	}

	if me.Type().IsTupleType() {
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
