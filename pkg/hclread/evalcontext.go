package hclread

import (
	"context"
	"encoding/base64"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/userfunc"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/walteh/terrors"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

func ExtractUserFuncs(ctx context.Context, ibdy hcl.Body, parent *hcl.EvalContext) (map[string]function.Function, hcl.Diagnostics) {
	userfuncs, _, diag := userfunc.DecodeUserFunctions(ibdy, "func", func() *hcl.EvalContext { return parent })
	if diag.HasErrors() {
		return nil, diag
	}

	return userfuncs, nil
}

func ExtractVariables(ctx context.Context, bdy *hclsyntax.Body, parent *hcl.EvalContext) (map[string]cty.Value, hcl.Diagnostics) {

	eectx := parent.NewChild()

	eectx.Functions = parent.Functions

	eectx.Variables = map[string]cty.Value{}

	// for _, v := range bdy.Attributes {
	// 	val, diag := v.Expr.Value(eectx)
	// 	if diag.HasErrors() {
	// 		return nil, diag
	// 	}
	// 	eectx.Variables[v.Name] = val
	// }

	// custvars := map[string]cty.Value{}

	// updateVariables := func(combos map[string][]cty.Value) {
	// 	for k, v := range combos {
	// 		if eectx.Variables[k] == cty.NilVal {
	// 			eectx.Variables[k] = cty.ObjectVal(map[string]cty.Value{})
	// 		}
	// 		wrk := eectx.Variables[k].AsValueMap()
	// 		if wrk == nil {
	// 			wrk = map[string]cty.Value{}
	// 		}

	// 		for _, v2 := range v {
	// 			for k2, v3 := range v2.AsValueMap() {
	// 				wrk[k2] = v3
	// 			}
	// 		}
	// 		eectx.Variables[k] = cty.ObjectVal(wrk)
	// 		combos[k] = nil
	// 	}

	// 	return
	// }
	var reevaluate func(blk *hclsyntax.Block) hcl.Diagnostics
	reevaluate = func(blk *hclsyntax.Block) hcl.Diagnostics {
		// combos := make(map[string][]cty.Value, 0)
		// key, blks, diags := NewUnknownBlockEvaluation(ctx, eectx, blk)
		// if key != "" {
		// 	if combos[key] == nil {
		// 		combos[key] = make([]cty.Value, 0)
		// 	}

		// 	combos[key] = append(combos[key], blks)
		// }
		// combos := make(map[string][]cty.Value, 0)

		// updateVariables(combos)
		// name := append([]string{blk.Type}, blk.Labels...)
		// return diags
		return hclsyntax.Walk(blk, &Walker{EvalContext: eectx, backlog: []hclsyntax.Node{}})
		// 	// switch node := node.(type) {
		// 	case *hclsyntax.Block:
		// 		diags := hcl.Diagnostics{}
		// 		// if node.Type == "gen" {
		// 		// 	return hcl.Diagnostics{}
		// 		// }

		// 		// key, blks, diag := NewUnknownBlockEvaluation(ctx, eectx, node)
		// 		// if key != "" {
		// 		// 	if combos[key] == nil {
		// 		// 		combos[key] = make([]cty.Value, 0)
		// 		// 	}

		// 		// 	combos[key] = append(combos[key], blks)
		// 		// }

		// 		for _, attr := range node.Body.Blocks {
		// 			rev, noded, diagd := reevaluate(attr)
		// 			if combos[rev] == nil {
		// 				combos[rev] = make([]cty.Value, 0)
		// 			}
		// 			combos[rev] = append(combos[rev], noded)
		// 			updateVariables(noded)
		// 			diags = append(diags, diagd...)
		// 		}

		// 		return diags
		// 	case *hclsyntax.ObjectConsExpr:

		// 	case *hclsyntax.Attribute:
		// 		val, diag := node.Expr.Value(eectx)
		// 		if diag.HasErrors() {
		// 			return diag
		// 		}
		// 		eectx.Variables[node.Name] = val
		// 		// case *hclsyntax.ObjectConsExpr:

		// 	}

		// 	return hcl.Diagnostics{}
		// })
	}

	retrys := bdy.Blocks
	prevRetrys := []*hclsyntax.Block{}
	lastDiags := hcl.Diagnostics{}
	start := true
	runs := 0
	// starts := 0
	for (len(retrys) > 0 && len(prevRetrys) > len(retrys)) || start {
		runs++
		start = false
		newRetrys := []*hclsyntax.Block{}

		diags := hcl.Diagnostics{}

		for _, v := range retrys {
			// if v.Type == "gen" {
			// 	continue
			// }

			diagd := reevaluate(v)
			if diagd.HasErrors() {
				diags = append(diags, diagd...)
				newRetrys = append(newRetrys, v)
			}
		}

		fmt.Println(runs, len(diags), len(lastDiags), len(prevRetrys), len(retrys))

		if len(diags) < len(lastDiags) {
			start = true
		}

		for _, x := range diags {
			fmt.Println(x)
		}

		prevRetrys = retrys
		retrys = newRetrys
		lastDiags = diags
	}

	return eectx.Variables, lastDiags

}

type Walker struct {
	EvalContext *hcl.EvalContext
	backlog     []hclsyntax.Node
	strs        []string
}

func applyblocklabels(v *hclsyntax.Block) []string {
	strz := []string{}
	backwardsLabels := make([]string, len(v.Labels))
	copy(backwardsLabels, v.Labels)
	slices.Reverse(backwardsLabels)

	strz = append(strz, backwardsLabels...)

	strz = append(strz, v.Type)
	return strz
}

func (me *Walker) Enter(node hclsyntax.Node) hcl.Diagnostics {
	var last hclsyntax.Node
	if len(me.backlog) > 0 {
		last = me.backlog[len(me.backlog)-1]
	}
	switch v := node.(type) {
	case *hclsyntax.Block:
		me.strs = append(me.strs, applyblocklabels(v)...)
		me.backlog = append(me.backlog, v)
	case *hclsyntax.Attribute:
		me.strs = append(me.strs, v.Name)
		me.backlog = append(me.backlog, v)
	case hclsyntax.Attributes:
		for _, v2 := range v {
			// fmt.Println(v2.Name)
			// fmt.Println(v2.Name, v2.Range(), last.Range())
			if last == nil || v2.Range().ContainsPos(last.Range().Start) {
				me.strs = append(me.strs, v2.Name)
			}
		}
		me.backlog = append(me.backlog, v)
	// case *hclsyntax.LiteralValueExpr:
	// 	me.strs = append(me.strs, v.Val.AsString())

	default:
		me.backlog = append(me.backlog, v)

		// pp.Println("enter unknown", reflect.TypeOf(v).String())
		// return hcl.Diagnostics{}
	}

	// me.backlog = append(me.backlog, node)

	// fmt.Println("enter", me.strs)
	return hcl.Diagnostics{}
}

func (me *Walker) Exit(node hclsyntax.Node) hcl.Diagnostics {
	// fmt.Println("exit", me.strs)

	me.backlog = me.backlog[:len(me.backlog)-1]

	strs := []string{}

	last := hclsyntax.Node(node)

	backwardsBacklog := make([]hclsyntax.Node, len(me.backlog))
	copy(backwardsBacklog, me.backlog)
	slices.Reverse(backwardsBacklog)

	var diag hcl.Diagnostics

	for _, v := range backwardsBacklog {

		switch v := v.(type) {
		case *hclsyntax.Block:

			strs = append(strs, applyblocklabels(v)...)
			last = v

		case hclsyntax.Attributes:
			for _, v2 := range v {
				if v2.Range().ContainsPos(last.Range().Start) {
					strs = append(strs, v2.Name)
				}
			}
			last = v
		case *hclsyntax.ObjectConsExpr:
			// valz := map[string]cty.Value{}
			for _, item := range v.Items {
				if item.ValueExpr.Range().ContainsPos(last.Range().Start) {
					vald, diagd := item.KeyExpr.Value(me.EvalContext)
					if diagd.HasErrors() {
						diag = append(diag, diagd...)
						break
					}
					strs = append(strs, vald.AsString())
				}

				// // strs = append(strs, vald.AsString())

				// val2, diag2 := item.ValueExpr.Value(me.EvalContext)
				// if diag2.HasErrors() {
				// 	diag = append(diag, diagd...)
				// }

				// // pp.Println(vald)

				// valz[vald.AsString()] = val2
			}

			// val = cty.ObjectVal(valz)
		default:
			// pp.Println("unknown", reflect.TypeOf(v).String())

			// fmt.Println("unknown", v, reflect.TypeOf(v))
			last = v
		}
	}
	val := cty.NilVal
	switch node := node.(type) {

	case *hclsyntax.LiteralValueExpr:
		val = node.Val
	case *hclsyntax.Attribute:
		strs = append(strs, node.Name)
		vald, diagd := node.Expr.Value(me.EvalContext)
		if diagd.HasErrors() {
			diag = append(diag, diagd...)
		} else {
			val = vald
		}
	default:
		// fmt.Println("unknownout", reflect.TypeOf(node).String(), strs)
		// return diag
		return diag

	}

	// for _, v := range strs {
	// 	val = cty.ObjectVal(map[string]cty.Value{
	// 		v: val,
	// 	})
	// }

	// fmt.Println(strs)

	// var ok bool

	// check := []string{"dir", "gotestsum-bin", "tasks", "data", "taskfile", "gen"}
	// if len(check) == len(strs) {
	// 	ok = true
	// 	for i := range strs {
	// 		if strs[i] != check[i] {
	// 			ok = false
	// 			break
	// 		}
	// 	}
	// 	if ok {
	// 		pp.Println(val)
	// 	}
	// } else {
	// 	ok = false
	// }

	// objval := cty.ObjectVal(me.EvalContext.Variables)

	// pp.Println(val)

	slices.Reverse(strs)

	fmt.Println(strs)

	merged := applyToNextedContext(context.TODO(), cty.ObjectVal(me.EvalContext.Variables), strs, val)

	me.EvalContext.Variables = merged.AsValueMap()

	return diag
}

func applyToNextedContext(ctx context.Context, mapd cty.Value, strs []string, val cty.Value) cty.Value {
	if len(strs) == 0 {
		// fmt.Println(val)
		return val
	} else {

		if mapd.Type().IsObjectType() {
			obj := mapd.AsValueMap()
			// var wrk map[string]cty.Value
			// if obj[strs[0]].IsNull() {
			// 	wrk = map[string]cty.Value{}
			// } else {
			// 	// fmt.Println(obj[strs[0]].Type().GoString())
			// 	if obj[strs[0]].CanIterateElements() {
			// 		wrk = obj[strs[0]].AsValueMap()
			// 	} else {
			// 		// wrk = cty.SetVal([]cty.Value{obj[strs[0]]})
			// 		wrk = map[string]cty.Value{}
			// 	}
			// }

			if obj == nil {
				obj = map[string]cty.Value{}
			}

			vald := applyToNextedContext(ctx, obj[strs[0]], strs[1:], val)

			obj[strs[0]] = vald

			return cty.ObjectVal(obj)
		}

		if mapd.Type().IsListType() {
			obj := mapd.AsValueSet()
			obj.Add(val)
			return cty.ListVal(obj.Values())

			// vars := obj.Values()
			// vars = append(vars, applyToNextedContext(ctx, cty.Value{}, strs[1:], val))
		}

		fmt.Println("outtttt", mapd.Type().GoString(), strs)

		obj := map[string]cty.Value{}

		vald := applyToNextedContext(ctx, obj[strs[0]], strs[1:], val)

		obj[strs[0]] = vald

		return cty.ObjectVal(obj)

		// if mapd.Type().
		// return val

		// panic("ahh " + mapd.Type().GoString())

		// var wrk map[string]cty.Value
		// if mapd[strs[0]].IsNull() {
		// 	wrk = map[string]cty.Value{}
		// } else {
		// 	// fmt.Println(mapd[strs[0]].Type().GoString())
		// 	if mapd[strs[0]].CanIterateElements() {
		// 		wrk = mapd[strs[0]].AsValueMap()
		// 	} else {
		// 		// wrk = cty.SetVal([]cty.Value{mapd[strs[0]]})
		// 		wrk = map[string]cty.Value{}
		// 	}
		// }

		// vald := applyToNextedContext(ctx, wrk, strs[1:], val)

		// mapd[strs[0]] = vald

		// return cty.ObjectVal(mapd)
	}
}

const MetaKey = "____meta"

func NewUnknownBlockEvaluation(ctx context.Context, ectx *hcl.EvalContext, block *hclsyntax.Block) (key string, res cty.Value, diags hcl.Diagnostics) {

	tmp := make(map[string]cty.Value)

	for _, attr := range block.Body.Attributes {
		// Evaluate the attribute's expression to get a cty.Value
		val, err := attr.Expr.Value(ectx)
		if err.HasErrors() {
			return "", cty.Value{}, err
		}

		tmp[attr.Name] = val
	}

	meta := map[string]cty.Value{
		"label": cty.StringVal(strings.Join(block.Labels, ".")),
	}

	tmp[MetaKey] = cty.ObjectVal(meta)

	for _, blkd := range block.Body.Blocks {

		key, blks, diags := NewUnknownBlockEvaluation(ctx, ectx, blkd)
		if diags.HasErrors() {
			return "", cty.Value{}, diags
		}

		if tmp[key] == cty.NilVal {
			tmp[key] = cty.ObjectVal(map[string]cty.Value{})
		}

		wrk := tmp[key].AsValueMap()
		if wrk == nil {
			wrk = map[string]cty.Value{}
		}

		for k, v := range blks.AsValueMap() {
			wrk[k] = v
		}

		tmp[key] = cty.ObjectVal(wrk)
	}

	for _, lab := range block.Labels {
		tmp = map[string]cty.Value{
			lab: cty.ObjectVal(tmp),
		}
	}

	return block.Type, cty.ObjectVal(tmp), hcl.Diagnostics{}

}

func NewContextFromFile(ctx context.Context, fle []byte, name string) (*hcl.File, *hcl.EvalContext, *hclsyntax.Body, hcl.Diagnostics, error) {

	hcldata, errd := hclsyntax.ParseConfig(fle, name, hcl.InitialPos)
	if errd.HasErrors() {
		return nil, nil, nil, errd, nil
	}

	ectx := &hcl.EvalContext{
		Functions: NewFunctionMap(),
		Variables: map[string]cty.Value{},
	}

	// will always work
	bdy := hcldata.Body.(*hclsyntax.Body)

	// process funcs
	funcs, diag := ExtractUserFuncs(ctx, bdy, ectx)
	if diag.HasErrors() {
		return nil, nil, nil, diag, nil
	}

	for k, v := range funcs {
		ectx.Functions[k] = v
	}

	// todo, do we need to remove the func blocks from the body?

	// process variables
	vars, diag := ExtractVariables(ctx, bdy, ectx)
	if diag.HasErrors() {
		return nil, nil, nil, diag, nil
	}

	for k, v := range vars {
		ectx.Variables[k] = v
	}

	return hcldata, ectx, bdy, nil, nil
}

func NewFunctionMap() map[string]function.Function {

	return map[string]function.Function{
		"jsonencode": stdlib.JSONEncodeFunc,
		"jsondecode": stdlib.JSONDecodeFunc,
		"csvdecode":  stdlib.CSVDecodeFunc,
		// "yamlencode": stdlib.YAMLDecodeFunc,
		"equal":      stdlib.EqualFunc,
		"notequal":   stdlib.NotEqualFunc,
		"concat":     stdlib.ConcatFunc,
		"format":     stdlib.FormatFunc,
		"join":       stdlib.JoinFunc,
		"lower":      stdlib.LowerFunc,
		"upper":      stdlib.UpperFunc,
		"replace":    stdlib.ReplaceFunc,
		"split":      stdlib.SplitFunc,
		"substr":     stdlib.SubstrFunc,
		"trimprefix": stdlib.TrimPrefixFunc,
		"trimspace":  stdlib.TrimSpaceFunc,
		"trimsuffix": stdlib.TrimSuffixFunc,
		"chomp":      stdlib.ChompFunc,
		"label": function.New(&function.Spec{
			Description: `Gets the label of an hcl block`,
			Params: []function.Parameter{
				{
					Name:             "block",
					Type:             cty.DynamicPseudoType,
					AllowUnknown:     true,
					AllowDynamicType: true,
					AllowNull:        false,
					AllowMarked:      true,
				},
			},
			Type: function.StaticReturnType(cty.String),
			Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
				if len(args) != 1 {
					return cty.NilVal, terrors.Errorf("expected 1 argument, got %d", len(args))
				}

				mp := args[0].AsValueMap()
				if mp == nil {
					return cty.NilVal, terrors.Errorf("expected map, got %s", args[0].GoString())
				}

				if mp[MetaKey] == cty.NilVal {
					return cty.NilVal, terrors.Errorf("expected map with _label, got %s", args[0].GoString())
				}

				mp = mp[MetaKey].AsValueMap()
				if mp == nil {
					return cty.NilVal, terrors.Errorf("expected map with _label, got %s", args[0].GoString())
				}

				return cty.StringVal(mp["label"].AsString()), nil
			},
		}),
		"base64encode": function.New(&function.Spec{
			Description: `Returns the Base64-encoded version of the given string.`,
			Params: []function.Parameter{
				{
					Name:             "str",
					Type:             cty.String,
					AllowUnknown:     false,
					AllowDynamicType: false,
					AllowNull:        false,
				},
			},
			Type: function.StaticReturnType(cty.String),
			// RefineResult: refineNonNull,
			Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
				if len(args) != 1 {
					return cty.NilVal, terrors.Errorf("expected 1 argument, got %d", len(args))
				}
				if args[0].IsNull() {
					return cty.StringVal(""), nil
				}

				if args[0].Type() != cty.String {
					return cty.NilVal, terrors.Errorf("expected string, got %s", args[0].GoString())
				}
				return cty.StringVal(base64.StdEncoding.EncodeToString([]byte(args[0].AsString()))), nil
			},
		}),
		"base64decode": function.New(&function.Spec{
			Description: `Returns the Base64-decoded version of the given string.`,
			Params: []function.Parameter{
				{
					Name:             "str",
					Type:             cty.String,
					AllowUnknown:     false,
					AllowDynamicType: false,
					AllowNull:        false,
				},
			},
			Type: function.StaticReturnType(cty.String),
			// RefineResult: refineNonNull,
			Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
				if len(args) != 1 {
					return cty.NilVal, terrors.Errorf("expected 1 argument, got %d", len(args))
				}
				if args[0].IsNull() {
					return cty.StringVal(""), nil
				}
				if args[0].Type() != cty.String {
					return cty.NilVal, terrors.Errorf("expected string, got %s", args[0].GoString())
				}
				dec, err := base64.StdEncoding.DecodeString(args[0].AsString())
				if err != nil {
					return cty.NilVal, err
				}
				return cty.StringVal(string(dec)), nil
			},
		}),
	}
}

// hclsyntax.Walk(bdy, func(n hclsyntax.Node) (hcl.Diagnostics, hclsyntax.Node) {
// 	switch n := n.(type) {
// 	case *hclsyntax.Attribute:
// 		if n.Name == "var" {
// 			return hcl.Diagnostics{}, n
// 		}
// 	case *hclsyntax.ForExpr:
// 		return hcl.Diagnostics{}, n
// 	case *hclsyntax.FunctionCallExpr:
// 		if n.Name == "var" {
// 			return hcl.Diagnostics{}, n
// 		}
// 	case *hclsyntax.IndexExpr:
// 		if n.Name == "var" {
// 			return hcl.Diagnostics{}, n
// 		}
// 	case *hclsyntax.ObjectConsExpr:

// 		for _, item := range n.Items {

// 			if item.KeyExpr.(*hclsyntax.LiteralValueExpr).Val.(*cty.StringVal).AsString() == "var" {
// 				return hcl.Diagnostics{}, n
// 			}
// 		}
// 	case *hclsyntax.ScopeTraversalExpr:
// 		if n.Traversal[0].(hclsyntax.TraverseAttr).Name == "var" {
// 			return hcl.Diagnostics{}, n
// 		}
// 	case *hclsyntax.TemplateExpr:
// 		return hcl.Diagnostics{}, n
// 	case *hclsyntax.TemplateWrapExpr:
// 		return hcl.Diagnostics{}, n
// 	case *hclsyntax.TupleConsExpr:

// 		for _, item := range n.Exprs {

// 			if item.(*hclsyntax.LiteralValueExpr).Val.(*cty.StringVal).AsString() == "var" {
// 				return hcl.Diagnostics{}, n
// 			}
// 		}
// 	case *hclsyntax.UnaryOpExpr:
// 		if n.Op == "var" {
// 			return hcl.Diagnostics{}, n
// 		}
// 	}
