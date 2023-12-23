package hclread

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-faster/errors"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/walteh/retab/schemas"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

// var rootScheme = &hcl.BodySchema{
// 	Blocks: []hcl.BlockHeaderSchema{
// 		{
// 			Type:       "file",
// 			LabelNames: []string{"name"},
// 		},
// 		{
// 			Type: "func",
// 		},
// 	},
// }

// var fileScheme = &hcl.BodySchema{
// 	Attributes: []hcl.AttributeSchema{
// 		{
// 			Name:     "dir",
// 			Required: true,
// 		},
// 		{
// 			Name:     "schema",
// 			Required: true,
// 		},
// 		{
// 			Name:     "data",
// 			Required: true,
// 		},
// 	},
// }

var (
// _ ValidationBlock = (*FileBlockEvaluation)(nil)
// _ ValidationBlock = (*AnyBlockEvaluation)(nil)
)

type AnyBlockEvaluation struct {
	Name       string
	Content    map[string]cty.Value
	Validation []*ValidationError
}

type FileBlockEvaluation struct {
	Name   string
	Schema string
	Dir    string
	// ContentRaw cty.Value
	Content    map[string]any
	Validation []*ValidationError
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

	return schema.Validate(me.Content)
}

func (me *FileBlockEvaluation) ValidateJSONSchemaProperty(ctx context.Context, prop string) error {
	schema, err := me.GetProperties(ctx)
	if err != nil {
		return err
	}

	if schema[prop] == nil {
		return errors.Errorf("property %q not found", prop)
	}

	return schema[prop].Validate(me.Content)

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
	File  *FileBlockEvaluation
	Other []*AnyBlockEvaluation
}

func NewFullEvaluation(ctx context.Context, ectx *hcl.EvalContext, file *hclsyntax.Body) (res *FullEvaluation, err error) {

	var fle *FileBlockEvaluation

	other := make([]*AnyBlockEvaluation, 0)

	for _, block := range file.Blocks {
		switch block.Type {
		case "file":
			blk, err := NewFileBlockEvaluation(ctx, ectx, block)
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
				lerr, ok := LoadValidationErrors(ctx, attr.Expr, ectx, err)
				if ok {
					continue
				}
				fle.Validation = append(fle.Validation, lerr...)
			}
		}
	}

	return &FullEvaluation{
		File:  fle,
		Other: other,
	}, nil
}

func (me *FileBlockEvaluation) PropertyEvaluation(ctx context.Context, ectx *hcl.EvalContext, block *hclsyntax.Block) ([]*ValidationError, error) {

	schema, err := me.GetProperties(ctx)
	if err != nil {
		return nil, err
	}

	errs := make([]*ValidationError, 0)

	if schema[block.Labels[0]] == nil {
		// return nil, errors.Errorf("property %q not found", schema[block.Labels[0]])
		return errs, nil
	}

	for _, attr := range block.Body.Attributes {
		vld, ok := LoadValidationErrors(ctx, attr.Expr, ectx, nil)
		if !ok {
			continue
		}

		if vld != nil {
			errs = append(errs, vld...)
		}
	}

	return errs, nil

}

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

func NewFileBlockEvaluation(ctx context.Context, ectx *hcl.EvalContext, block *hclsyntax.Block) (res *FileBlockEvaluation, err error) {

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
			for _, diag := range err.Errs() {
				fmt.Println(diag)
			}
			return nil, errors.Wrapf(err, "failed to evaluate %q", attr.Name)
		}

		switch attr.Name {
		case "dir":
			blk.Dir = val.AsString()
		case "schema":
			blk.Schema = val.AsString()
		case "data":
			wrk, err := stdlib.JSONEncode(val)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to encode %q", attr.Name)
			}

			err = json.Unmarshal([]byte(wrk.AsString()), &blk.Content)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to decode %q", attr.Name)
			}

			dataAttr = attr.Expr

		default:
			// ignore unknown attributes
			continue
		}

	}

	// Validate the block body against the schema
	if errv := blk.ValidateJSONSchema(ctx); errv != nil {
		if lerr, ok := LoadValidationErrors(ctx, dataAttr, ectx, errv); ok {
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
