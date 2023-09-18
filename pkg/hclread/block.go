package hclread

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

var rootScheme = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "file",
			LabelNames: []string{"name"},
		},
	},
}

var fileScheme = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "dir",
			Required: true,
		},
		{
			Name:     "schema",
			Required: true,
		},
		{
			Name:     "data",
			Required: true,
		},
	},
}

type Block struct {
	Name   string
	Schema string
	Dir    string
	// ContentRaw cty.Value
	Content map[string]any
}

func (me *Block) GetJSONSchema(ctx context.Context) (*jsonschema.Schema, error) {
	return LoadJsonSchemaFile(ctx, me.Schema)
}

func (me *Block) ValidateJSONSchema(ctx context.Context) error {
	schema, err := me.GetJSONSchema(ctx)
	if err != nil {
		return err
	}

	return schema.Validate(me.Content)
}

func ParseBlocksFromFile(ctx context.Context, fle afero.File) (res []*Block, err hcl.Diagnostics) {

	all, errd := afero.ReadAll(fle)
	if errd != nil {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to read file",
			Detail:   err.Error(),
		}}
	}

	hcldata, err := hclparse.NewParser().ParseHCL(all, fle.Name())
	if err.HasErrors() {
		return nil, err
	}

	ctn, err := hcldata.Body.Content(rootScheme)
	if err.HasErrors() {
		return nil, err
	}

	var blocks []*Block
	for _, block := range ctn.Blocks {
		switch block.Type {
		case "file":
			blk := &Block{
				Name: block.Labels[0],
			}
			ctn2, err := block.Body.Content(fileScheme)
			if err.HasErrors() {
				return nil, err
			}
			for _, attr := range ctn2.Attributes {
				evalContext := &hcl.EvalContext{}

				// Evaluate the attribute's expression to get a cty.Value
				val, err := attr.Expr.Value(evalContext)
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
						return nil, hcl.Diagnostics{
							&hcl.Diagnostic{
								Severity: hcl.DiagError,
								Summary:  "Failed to encode data",
								Detail:   err.Error(),
							},
						}
					}

					err = json.Unmarshal([]byte(wrk.AsString()), &blk.Content)
					if err != nil {
						return nil, hcl.Diagnostics{
							&hcl.Diagnostic{
								Severity: hcl.DiagError,
								Summary:  "Failed to decode data",
								Detail:   err.Error(),
							},
						}
					}
				default:
					// ignore unknown attributes
					continue
				}
			}

			blocks = append(blocks, blk)
		}
	}

	return blocks, nil
}
