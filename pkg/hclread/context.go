package hclread

import (
	"context"
	"slices"
	"strconv"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

type SudoContext struct {
	ParentKey string
	Parent    *SudoContext
	Map       map[string]*SudoContext
	Array     []*SudoContext
	Value     *cty.Value
	UserFuncs map[string]function.Function
}

func (me *SudoContext) ApplyValue(val cty.Value) {
	if val.Type().IsObjectType() {
		me.ApplyValueMap(val.AsValueMap())
	} else {
		me.Value = &val
	}
}

func (me *SudoContext) ApplyKeyVal(key string, val cty.Value) {
	// if me.ParentKey == ArrKey {
	// 	me.Parent.ApplyValue(me.ToValue())
	// } else {
	me.NewChild(key).ApplyValue(val)
	// }
}

func (me *SudoContext) ApplyValueMap(val map[string]cty.Value) {
	for k, v := range val {
		me.ApplyKeyVal(k, v)
	}
}

func (me *SudoContext) ApplyArray(arr []cty.Value) {
	if me.ParentKey != ArrKey {
		me.ApplyValue(cty.TupleVal(arr))
		return
	}
	for _, v := range arr {
		me.Array = append(me.Array, &SudoContext{
			Parent: me,
			Value:  &v,
		})
	}
}

func (parent *SudoContext) ApplyBody(ctx context.Context, body *hclsyntax.Body) hcl.Diagnostics {
	return ExtractVariables(ctx, body, parent)
}

func (me *SudoContext) ToValue() cty.Value {
	if me.Value != nil {
		return *me.Value
	}

	if me.ParentKey == ArrKey {
		vars := me.BuildStaticVarsList()
		return cty.TupleVal(vars)
	}

	if me.Array != nil {
		arr := make([]cty.Value, len(me.Array))
		for i, v := range me.Array {
			arr[i] = v.ToValue()
		}
		return cty.TupleVal(arr)
	}

	obj := make(map[string]cty.Value, len(me.Map))
	for k, v := range me.Map {
		obj[k] = v.ToValue()
	}
	return cty.ObjectVal(obj)
}

func (me *SudoContext) BuildStaticEvalVars() map[string]cty.Value {

	wrk := map[string]cty.Value{}
	for k, v := range me.Map {
		wrk[k] = v.ToValue()
	}

	return wrk

}

func (me *SudoContext) BuildStaticVarsList() []cty.Value {

	type sorter struct {
		Key   string
		Value cty.Value
	}

	wrk := make([]sorter, 0, len(me.Map))
	for srt, v := range me.Map {
		if v.Value != nil {
			wrk = append(wrk, sorter{
				Key:   srt,
				Value: *v.Value,
			})
		} else {
			wrk = append(wrk, sorter{
				Key:   srt,
				Value: cty.ObjectVal(v.BuildStaticEvalVars()),
			})
		}
	}

	slices.SortFunc(wrk, func(a, b sorter) int {
		intr, _ := strconv.Atoi(a.Key)
		intr2, _ := strconv.Atoi(b.Key)
		return intr - intr2
	})

	vals := make([]cty.Value, len(wrk))
	for i, v := range wrk {
		vals[i] = v.Value
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

func (me *SudoContext) NewNestedChild(key ...string) *SudoContext {
	wrk := me
	for _, v := range key {
		wrk = wrk.NewChild(v)
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

	internalParent = internalParent[file].AsValueMap()

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

	for k, v := range NewContextualizedFunctionMap(wc.Root().BuildStaticEvalVars(), file) {
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
