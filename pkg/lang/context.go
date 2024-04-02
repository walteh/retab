package lang

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/walteh/terrors"
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

func (me *SudoContext) IdentifyChild(strs ...string) *SudoContext {
	wrk := me
	for _, v := range strs {
		wrkd, ok := wrk.Map[v]
		if !ok {
			return nil
		}
		wrk = wrkd
	}
	return wrk
}

type RemappableSudoContextArray []*SudoContext

func (me *SudoContext) ApplyValue(met cty.Value) {
	me.Value = &met
}

func (me *SudoContext) ApplyKeyVal(key string, val cty.Value, r hcl.Range) hcl.Diagnostics {
	child, err := me.NewNonBlockChild(key, r)
	if err != nil {
		return hcl.Diagnostics{{
			Severity: hcl.DiagError,
			Summary:  "unable to apply key value, key already exists as block",
			Detail:   err.Error(),
			Subject:  &r,
		}}
	}
	child.ApplyValue(val)
	return nil
}

func (parent *SudoContext) ApplyBody(ctx context.Context, body *hclsyntax.Body) hcl.Diagnostics {
	return ExtractVariables(ctx, body, parent)
}

func (me *SudoContext) ToValueWithExtraContext() cty.Value {

	var val cty.Value

	if me.Value != nil {
		val = me.ToValue()
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

	if _, ok := me.Meta.(*IncomleteBlockMeta); ok {
		val = val.Mark(isIncompleteBlock{})
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

func (me *SudoContext) NewBlockLabelChild(label string, rnge hcl.Range) (*SudoContext, error) {
	wrk := me.NewChild(label, rnge)
	_, isBlock := wrk.Meta.(*BlockLabelMeta)
	isNew := wrk.Meta.Range().String() == rnge.String()

	if !isNew && !isBlock {
		return nil, terrors.Errorf("block %q already exists at %s - cant create at %s", label, wrk.Meta.Range().String(), rnge.String())
	}

	wrk.Meta = &BlockLabelMeta{rnge}

	return wrk, nil
}

func (me *SudoContext) NewBlockChild(typ *hclsyntax.Block) (*SudoContext, error) {
	wrk, err := me.NewBlockLabelChild(typ.Type, typ.TypeRange)
	if err != nil {
		return nil, err
	}

	for i, v := range typ.LabelRanges {
		wrk = wrk.NewChild(typ.Labels[i], v)
		wrk.Meta = &BlockLabelMeta{v}
	}

	for _, v := range typ.Body.Blocks {
		_, err := wrk.NewBlockChild(v)
		if err != nil {
			return nil, err
		}
	}

	wrk.Meta = &IncomleteBlockMeta{typ}

	return wrk, nil
}

func (me *SudoContext) NewNonBlockChild(key string, rnge hcl.Range) (*SudoContext, error) {
	wrk := me.NewChild(key, rnge)
	if _, ok := wrk.Meta.(*BlockLabelMeta); ok {
		return nil, terrors.Errorf("block %q already exists at %s - cant create at %s", key, wrk.Meta.Range().String(), rnge.String())
	}

	return wrk, nil
}

func (me *SudoContext) NewChild(key string, rnge hcl.Range) *SudoContext {

	if me.Map[key] != nil {

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

func (wc *SudoContext) GetAllUserFuncs() map[string]function.Function {
	mine := map[string]function.Function{}
	if wc.Parent != nil {
		for k, v := range wc.Parent.GetAllUserFuncs() {
			mine[k] = v
		}
	}
	for k, v := range wc.UserFuncs {
		mine[k] = v
	}
	return mine
}

func (wc *SudoContext) BuildStaticEvalContextWithFileData(file string) *hcl.EvalContext {

	wrkd := wc.Root().Map[FilesKey].Map[sanitizeFileName(file)]
	if wrkd == nil {
		panic(fmt.Sprintf("file %s not found", file))
	}
	wrk := wrkd.BuildStaticEvalContext()

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

	for k, v := range wc.GetAllUserFuncs() {
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

func (wc *SudoContext) ParentBlockMeta() BlockMeta {
	if wc.Parent == nil {
		return nil
	}

	if bm, ok := wc.Meta.(BlockMeta); ok {
		return bm
	}

	return wc.Parent.ParentBlockMeta()
}

func (me *SudoContext) BlocksOfType(name string) ([]*SudoContext, error) {
	blks := []*SudoContext{}
	for k, v := range me.Map {
		if k == name {
			for _, d := range v.Map {
				for _, ok := d.Meta.(*BlockLabelMeta); ok; _, ok = d.Meta.(*BlockLabelMeta) {
					d = d.List()[0]
				}
				if bm, ok := d.Meta.(BlockMeta); ok {
					// fmt.Println(name, bm.Block().Labels, reflect.TypeOf(bm).String())
					if bm.Block().Type == name {
						blks = append(blks, d)
					}
				} else {
					return nil, terrors.Errorf("unable to find blocks  of type %q - item at %s is not a block", name, v.Meta.Range().String())
				}
			}
		}
	}
	return blks, nil
}

func (me *SudoContext) GetAllFileLevelBlocksOfType(name string) ([]*SudoContext, error) {
	files := me.Map[FilesKey].Map
	out := []*SudoContext{}
	for _, v := range files {
		fleblks, err := v.BlocksOfType(name)
		if err != nil {
			return nil, err
		}
		out = append(out, fleblks...)
	}
	return out, nil
}

func (me *SudoContext) Match(keys []string, val cty.Value) (bool, hcl.Diagnostics) {
	vv := me.ToValueWithExtraContext()
	for _, k := range keys {
		vv = vv.GetAttr(k)
		if vv == cty.NilVal {
			return false, nil
		}
	}

	if val.Type() != cty.String {
		return false, nil
	}

	reg, err := regexp.Compile(val.AsString())
	if err != nil {
		return false, hcl.Diagnostics{{
			Severity: hcl.DiagError,
			Summary:  "unable to compile regex",
			Detail:   err.Error(),
			Subject:  rangeOf(val).Ptr(),
		}}
	}

	vv, _ = vv.Unmark()

	return reg.MatchString(vv.AsString()), nil

}

func FilterSudoContextWithRegex(sctx []*SudoContext, keys []string, reg cty.Value) ([]*SudoContext, error) {
	filtered := make([]*SudoContext, 0)
	for _, vv := range sctx {
		mok, err := vv.Match(keys, reg)
		if err != nil {
			return nil, err
		}
		if mok {
			filtered = append(filtered, vv)
		}
	}

	return filtered, nil
}
