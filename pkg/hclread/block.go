package hclread

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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
	Validation *ValidationError
}

func (me *BlockEvaluation) GetJSONSchema(ctx context.Context) (*jsonschema.Schema, error) {
	return schemas.LoadJSONSchema(ctx, me.Schema)
}

func (me *BlockEvaluation) ValidateJSONSchema(ctx context.Context) error {
	schema, err := me.GetJSONSchema(ctx)
	if err != nil {
		return err
	}

	return schema.Validate(me.Content)
}

func NewBlockEvaluation(ctx context.Context, ectx *hcl.EvalContext, block *hclsyntax.Block) (res *BlockEvaluation, err error) {
	switch block.Type {
	case "file":
		blk := &BlockEvaluation{
			Name: block.Labels[0],
		}

		var dataAttr hcl.Expression

		for _, attr := range block.Body.Attributes {

			// Evaluate the attribute's expression to get a cty.Value
			val, err := attr.Expr.Value(ectx)
			if err.HasErrors() {
				return nil, err
			}

			switch attr.Name {
			case "dir":
				blk.Dir = val.AsString()
			case "schema":
				blk.Schema = val.AsString()
			case "data":
				wrk, err := stdlib.JSONEncode(val)
				if err != nil {
					return nil, err
				}

				err = json.Unmarshal([]byte(wrk.AsString()), &blk.Content)
				if err != nil {
					return nil, err
				}

				dataAttr = attr.Expr

			default:
				// ignore unknown attributes
				continue
			}

		}

		// Validate the block body against the schema
		if errv := blk.ValidateJSONSchema(ctx); errv != nil {
			if lerr, err := LoadValidationErrors(ctx, dataAttr, errv); err != nil {
				return nil, err
			} else {
				blk.Validation = lerr
			}
		}

		return blk, nil

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
