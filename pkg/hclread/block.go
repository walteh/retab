package hclread

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/go-faster/errors"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/walteh/retab/schemas"
	"github.com/walteh/terrors"
	"github.com/walteh/yaml"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

type AnyBlockEvaluation struct {
	Name       string
	Content    map[string]cty.Value
	Validation []*ValidationError
}

type FileBlockEvaluation struct {
	Name          string
	Schema        string
	Dir           string
	OrderedOutput yaml.MapSlice
	RawOutput     any
	Validation    []*ValidationError
}

func (me *FileBlockEvaluation) GetJSONSchema(ctx context.Context) (*jsonschema.Schema, error) {
	return schemas.LoadJSONSchema(ctx, me.Schema)
}

func (me *FileBlockEvaluation) GetProperties(ctx context.Context) (map[string]*jsonschema.Schema, error) {
	schema, err := me.GetJSONSchema(ctx)
	if err != nil {
		return nil, err
	}

	m := make(map[string]*jsonschema.Schema)

	getAllDefs("root", schema, m)

	return m, nil
}

func (me *FileBlockEvaluation) ValidateJSONSchema(ctx context.Context) error {
	schema, err := me.GetJSONSchema(ctx)
	if err != nil {
		return err
	}

	return schema.Validate(me.RawOutput)
}

func (me *FileBlockEvaluation) ValidateJSONSchemaProperty(ctx context.Context, prop string) error {
	if prop == MetaKey {
		return nil
	}

	schema, err := me.GetProperties(ctx)
	if err != nil {
		return err
	}

	if schema[prop] == nil {
		return errors.Errorf("property %q not found", prop)
	}

	return schema[prop].Validate(me.RawOutput)

}

func getAllDefs(parent string, schema *jsonschema.Schema, defs map[string]*jsonschema.Schema) {
	////////////////////////////////////////
	// not sure if this is needed or not
	if defs[parent] != nil {
		return
	}
	////////////////////////////////////////

	if schema == nil {
		return
	}

	for k, v := range schema.DependentSchemas {
		if ok := defs[k]; ok != nil {
			continue
		}
		defs[k] = v
		getAllDefs(k, v, defs)
	}

	for k, v := range schema.Properties {
		if ok := defs[k]; ok != nil {
			continue
		}
		defs[k] = v
		getAllDefs(k, v, defs)
	}

	for _, v := range schema.AllOf {
		getAllDefs(parent, v, defs)
	}

	for _, v := range schema.AnyOf {
		getAllDefs(parent, v, defs)
	}

	for _, v := range schema.OneOf {
		getAllDefs(parent, v, defs)
	}

	for _, v := range schema.PatternProperties {
		getAllDefs(parent, v, defs)
	}

	switch v := schema.Items.(type) {
	case *jsonschema.Schema:
		getAllDefs(parent, v, defs)
	case []*jsonschema.Schema:
		for _, v := range v {
			getAllDefs(parent, v, defs)
		}
	}
}

type FullEvaluation struct {
	File   *FileBlockEvaluation
	Other  []*AnyBlockEvaluation
	Source string
}

func NewFullEvaluation(ctx context.Context, ectx *hcl.EvalContext, file *hclsyntax.Body, preserveOrder bool, source string) (res *FullEvaluation, err error) {

	var fle *FileBlockEvaluation

	other := make([]*AnyBlockEvaluation, 0)

	for _, block := range file.Blocks {
		switch block.Type {
		case "file":
			blk, err := NewFileBlockEvaluation(ctx, ectx, block, preserveOrder)
			if err != nil {
				return nil, err
			}

			fle = blk
			break
		default:
			// blk, err := NewAnyBlockEvaluation(ctx, ectx, block)
			// if err != nil {
			// 	return nil, err
			// }
			// other = append(other, blk)
		}

	}

	if fle == nil {
		return nil, errors.Errorf("no file block found")
	}

	for _, block := range file.Blocks {
		if block.Type == "file" {
			continue
		}

		for _, attr := range block.Body.Attributes {
			if attr.Name != "file" {
				continue
			}

			err := fle.ValidateJSONSchemaProperty(ctx, block.Type)
			if err != nil {
				lerr, err := LoadValidationErrors(ctx, attr.Expr, ectx, err)
				if err != nil {
					return nil, err
				}
				fle.Validation = append(fle.Validation, lerr...)
			}
		}
	}

	return &FullEvaluation{
		File:   fle,
		Other:  other,
		Source: source,
	}, nil
}

// func (me *FileBlockEvaluation) PropertyEvaluation(ctx context.Context, ectx *hcl.EvalContext, block *hclsyntax.Block) ([]*ValidationError, error) {

// 	schema, err := me.GetProperties(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	errs := make([]*ValidationError, 0)

// 	if schema[block.Labels[0]] == nil {
// 		// return nil, errors.Errorf("property %q not found", schema[block.Labels[0]])
// 		return errs, nil
// 	}

// 	for _, attr := range block.Body.Attributes {
// 		vld, err := LoadValidationErrors(ctx, attr.Expr, ectx, nil)
// 		if err != nil {
// 			return nil, err
// 		}

// 		if vld != nil {
// 			errs = append(errs, vld...)
// 		}
// 	}

// 	return errs, nil

// }

const MetaKey = "____meta"

func NewAnyBlockEvaluation(ctx context.Context, ectx *hcl.EvalContext, block *hclsyntax.Block) (key string, res cty.Value, err error) {

	tmp := make(map[string]cty.Value)

	for _, attr := range block.Body.Attributes {
		// Evaluate the attribute's expression to get a cty.Value
		val, err := attr.Expr.Value(ectx)
		if err.HasErrors() {
			return "", cty.Value{}, errors.Wrapf(err, "failed to evaluate %q", attr.Name)
		}

		tmp[attr.Name] = val
	}

	meta := map[string]cty.Value{
		"label": cty.StringVal(strings.Join(block.Labels, ".")),
	}

	tmp[MetaKey] = cty.ObjectVal(meta)

	for _, blkd := range block.Body.Blocks {

		key, blks, err := NewAnyBlockEvaluation(ctx, ectx, blkd)
		if err != nil {
			return "", cty.Value{}, err
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

	return block.Type, cty.ObjectVal(tmp), nil

}

func roll(e hclsyntax.Expression, ectx *hcl.EvalContext) (any, error) {

	if x, ok := e.(*hclsyntax.ObjectConsExpr); ok {
		group := make(yaml.MapSlice, 0)
		for _, rr := range x.Items {
			kvf, err := rr.KeyExpr.Value(ectx)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to evaluate %q", rr.KeyExpr)
			}
			if kvf.AsString() == MetaKey {
				continue
			}
			rz, errd := roll(rr.ValueExpr, ectx)
			if errd != nil {
				return nil, errd
			}
			if rz == nil {
				continue
			}
			group = append(group, yaml.MapItem{Key: kvf.AsString(), Value: rz})
		}
		return group, nil
	} else if x, ok := e.(*hclsyntax.TupleConsExpr); ok {
		wrk := make([]any, 0)
		for _, exp := range x.Exprs {
			r, err := roll(exp, ectx)
			if err != nil {
				return nil, terrors.Wrapf(err, "failed to evaluate %q", exp)
			}
			if r == nil {
				continue
			}
			wrk = append(wrk, r)
		}
		return wrk, nil
	} else {
		evaled, errd := e.Value(ectx)
		if errd != nil {
			return nil, errors.Wrapf(errd, "failed to evaluate %q", e)
		}

		return noMetaJsonEncode(evaled)
	}
}

var reg1 = regexp.MustCompile(`,?\"` + MetaKey + `\":{.*?}`)
var reg2 = regexp.MustCompile(`{,`)

func noMetaJsonEncode(v cty.Value) (any, error) {
	ok, err := stdlib.JSONEncode(v)
	if err != nil {
		return nil, err
	}
	oks := reg1.ReplaceAllString(ok.AsString(), "")
	oks = reg2.ReplaceAllString(oks, "{")
	ok = cty.StringVal(oks)

	var ok2 interface{}
	err = json.Unmarshal([]byte(ok.AsString()), &ok2)
	if err != nil {
		return nil, err
	}
	// if strings.Contains(ok.AsString(), MetaKey) {
	// 	fmt.Println("found meta, skipping", ok)
	// }
	return ok2, nil
}

func checkNestedForMeta(message string, validation []*jsonschema.ValidationError) bool {
	// if message == "additionalProperties '"+MetaKey+"' not allowed" {
	// 	return true
	// }
	// for _, v := range validation {
	// 	return checkNestedForMeta(v.Message, v.Causes)
	// }
	return false
}

func NewFileBlockEvaluation(ctx context.Context, ectx *hcl.EvalContext, block *hclsyntax.Block, preserveOrder bool) (res *FileBlockEvaluation, err error) {

	if block.Type != "file" {
		return nil, errors.Errorf("invalid block type %q", block.Type)
	}

	blk := &FileBlockEvaluation{
		Name: block.Labels[0],
	}

	var dataAttr hcl.Expression

	for _, attr := range block.Body.Attributes {

		// Evaluate the attribute's expression to get a cty.Value
		val, err := attr.Expr.Value(ectx)
		if err.HasErrors() {
			return nil, terrors.Wrapf(err, "failed to evaluate %q", attr.Name)
		}

		switch attr.Name {
		case "dir":
			blk.Dir = val.AsString()
		case "schema":
			blk.Schema = val.AsString()
		case "data":

			cnt := yaml.MapSlice{}

			slc, err := roll(attr.Expr, ectx)
			if err != nil {
				return nil, terrors.Wrap(err, "failed to evaluate")
			}

			if x, ok := slc.(yaml.MapSlice); ok {
				cnt = append(cnt, x...)
			}
			if x, ok := slc.(yaml.MapItem); ok {
				cnt = append(cnt, x)
			}

			mta, err := noMetaJsonEncode(val)
			if err != nil {
				return nil, err
			}

			blk.OrderedOutput = cnt

			blk.RawOutput = mta

			dataAttr = attr.Expr

		default:
			// ignore unknown attributes
			continue
		}
	}

	if dataAttr == nil {
		return nil, errors.Errorf("missing data attribute")
	}

	// Validate the block body against the schema
	if errv := blk.ValidateJSONSchema(ctx); errv != nil {
		if lerr, err := LoadValidationErrors(ctx, dataAttr, ectx, errv); err != nil {
			return nil, err
		} else {
			blk.Validation = lerr
		}
	}

	return blk, nil

}

type ValidationBlock interface {
	HasValidationErrors() bool
	Encode() ([]byte, error)
}

func (me *FileBlockEvaluation) HasValidationErrors() bool {
	return me.Validation != nil
}
