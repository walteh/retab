package hclread

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-faster/errors"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/k0kubun/pp/v3"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/walteh/retab/schemas"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	"gopkg.in/yaml.v3"
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

type BlockEvaluation struct {
	Name   string
	Schema string
	Dir    string
	// ContentRaw cty.Value
	Content    map[string]any
	Validation []*ValidationError
}

func (me *BlockEvaluation) GetJSONSchema(ctx context.Context) (*jsonschema.Schema, error) {
	return schemas.LoadJSONSchema(ctx, me.Schema)
}

func (me *BlockEvaluation) GetProperties(ctx context.Context) (map[string]*jsonschema.Schema, error) {
	schema, err := me.GetJSONSchema(ctx)
	if err != nil {
		return nil, err
	}

	m := make(map[string]*jsonschema.Schema)

	getAllDefs("root", schema, m)

	pp.Println(m)

	return m, nil
}

func (me *BlockEvaluation) ValidateJSONSchema(ctx context.Context) error {
	schema, err := me.GetJSONSchema(ctx)
	if err != nil {
		return err
	}

	return schema.Validate(me.Content)
}

func (me *BlockEvaluation) ValidateJSONSchemaProperty(ctx context.Context, prop string) error {
	schema, err := me.GetProperties(ctx)
	if err != nil {
		return err
	}

	if schema[prop] == nil {
		return fmt.Errorf("property %q not found", prop)
	}

	return schema[prop].Validate(me.Content)

}

func getAllDefs(parent string, schema *jsonschema.Schema, defs map[string]*jsonschema.Schema) {
	if schema.String() == "permissions" {
		fmt.Println(parent)
	}

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

func NewBlockEvaluation(ctx context.Context, ectx *hcl.EvalContext, block *hclsyntax.Block) (res *BlockEvaluation, err error) {
	switch block.Type {
	case "file":
		{
			blk := &BlockEvaluation{
				Name: block.Labels[0],
			}

			var dataAttr hcl.Expression

			for _, attr := range block.Body.Attributes {

				// Evaluate the attribute's expression to get a cty.Value
				val, err := attr.Expr.Value(ectx)
				if err.HasErrors() {
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
				if lerr, err := LoadValidationErrors(ctx, dataAttr, ectx, errv); err != nil {
					return nil, err
				} else {
					blk.Validation = lerr
				}
			}

			return blk, nil
		}
	default:

		return nil, fmt.Errorf("unknown block type %s", block.Type)
	}

}

func (me *BlockEvaluation) HasValidationErrors() bool {
	return me.Validation != nil
}

func (me *BlockEvaluation) Encode() ([]byte, error) {
	arr := strings.Split(me.Name, ".")
	if len(arr) < 2 {
		return nil, fmt.Errorf("invalid file name [%s] - missing extension", me.Name)
	}
	switch arr[len(arr)-1] {
	case "json":
		return json.MarshalIndent(me.Content, "", "\t")
	case "yaml":
		dat, err := yaml.Marshal(me.Content)
		if err != nil {
			return nil, err
		}

		return []byte(strings.ReplaceAll(string(dat), "\t", "")), nil
	// case "hcl":
	// 	return
	default:
		return nil, fmt.Errorf("unknown file extension [%s] in %s", arr[len(arr)-1], me.Name)
	}
}
