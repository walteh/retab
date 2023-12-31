package hclread

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/rs/zerolog"
	"github.com/walteh/retab/schemas"
	"github.com/walteh/terrors"
	"github.com/walteh/yaml"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

type AnyBlockEvaluation struct {
	Name    string
	Content map[string]cty.Value
	// Validation []*ValidationError
}

type FileBlockEvaluation struct {
	Name          string
	Schema        string
	Path          string
	OrderedOutput yaml.MapSlice
	RawOutput     any
	Source        string
	// Validation    []*ValidationError
}

func NewGenBlockEvaluation(ctx context.Context, ectx *hcl.EvalContext, file *hclsyntax.Body) (res *FileBlockEvaluation, diags hcl.Diagnostics, err error) {

	var fblock *hclsyntax.Block

	for _, block := range file.Blocks {
		switch block.Type {
		case "gen":
			fblock = block
			break
		default:
			// blk, err := NewAnyBlockEvaluation(ctx, ectx, block)
			// if err != nil {
			// 	return nil, err
			// }
			// other = append(other, blk)
		}
	}

	if fblock == nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "missing gen block",
				Detail:   "at least one gen block must be present",
				Subject:  file.Range().Ptr(),
			},
		}, nil
	}

	if len(fblock.Labels) != 1 {
		if len(fblock.Labels) == 0 {
			return nil, hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "missing block label",
					Detail:   "a gen block must have a label",
					Subject:  &fblock.TypeRange,
				},
			}, nil
		}
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "too many block labels",
				Detail:   "a gen block can only have one label",
				Subject:  &fblock.TypeRange,
			},
		}, nil
	}

	blk := &FileBlockEvaluation{
		Name:   fblock.Labels[0],
		Source: fblock.TypeRange.Filename,
	}

	var dataAttr hclsyntax.Expression

	attrs := []string{"path", "schema", "data"}

	for _, attrkey := range attrs {

		attr := fblock.Body.Attributes[attrkey]

		if attr == nil {
			if attrkey == "schema" {
				continue
			}
			return nil, hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "missing attribute",
				Detail:   "a file block must have a " + attrkey + " attribute",
				Subject:  &fblock.TypeRange,
			}}, nil
		}

		// Evaluate the attribute's expression to get a cty.Value
		val, diag := attr.Expr.Value(ectx)
		if diag.HasErrors() {
			return nil, diag, nil
		}

		switch attr.Name {
		case "path":
			blk.Path = val.AsString()

			blk.Path = sanatizeGenPath(blk.Path)

			ectx.Functions["ref"] = NewRefFunctionFromPath(ctx, blk.Path)
		case "schema":
			blk.Schema = val.AsString()
		case "data":

			cnt := yaml.MapSlice{}

			slc, diags, err := roll(attr.Expr, ectx)
			if err != nil || diags.HasErrors() {
				return nil, diags, err
			}

			if x, ok := slc.(yaml.MapSlice); ok {
				cnt = append(cnt, x...)
			}
			if x, ok := slc.(yaml.MapItem); ok {
				cnt = append(cnt, x)
			}

			mta, err := noMetaJsonEncode(val)
			if err != nil {
				return nil, hcl.Diagnostics{}, terrors.Wrap(err, "problem encoding json")
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
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "missing data attribute",
			Detail:   "a file block must have a data attribute",
			Subject:  &fblock.TypeRange,
		}}, nil
	}

	diags = hcl.Diagnostics{}

	if blk.Schema != "" {

		s, err := schemas.LoadJSONSchema(ctx, blk.Schema)
		if err != nil {
			return nil, hcl.Diagnostics{}, terrors.Wrap(err, "problem getting schema").Event(func(e *zerolog.Event) *zerolog.Event {
				return e.Int("schema_size", len(blk.Schema))
			})
		}

		// Validate the block body against the schema
		if errv := s.Validate(blk.RawOutput); errv != nil {
			if lerr, err := LoadValidationErrors(ctx, dataAttr, ectx, errv, file); err != nil {
				return nil, hcl.Diagnostics{}, terrors.Wrap(err, "problem loading validation errors")
			} else {
				for _, v := range lerr {
					diags = append(diags, v)
				}
			}
		}
	}

	return blk, diags, nil
}

func roll(e hclsyntax.Expression, ectx *hcl.EvalContext) (any, hcl.Diagnostics, error) {

	if x, ok := e.(*hclsyntax.ObjectConsExpr); ok {
		group := make(yaml.MapSlice, 0)
		for _, rr := range x.Items {
			kvf, diags := rr.KeyExpr.Value(ectx)
			if diags.HasErrors() {
				return nil, diags, nil
			}
			if kvf.AsString() == MetaKey {
				continue
			}
			rz, diags, errd := roll(rr.ValueExpr, ectx)
			if errd != nil || diags.HasErrors() {
				return nil, diags, errd
			}
			if rz == nil {
				continue
			}
			group = append(group, yaml.MapItem{Key: kvf.AsString(), Value: rz})
		}
		return group, hcl.Diagnostics{}, nil
	} else if x, ok := e.(*hclsyntax.TupleConsExpr); ok {
		wrk := make([]any, 0)
		for _, exp := range x.Exprs {
			r, diags, err := roll(exp, ectx)
			if err != nil || diags.HasErrors() {
				return nil, diags, err
			}
			if r == nil {
				continue
			}
			wrk = append(wrk, r)
		}
		return wrk, hcl.Diagnostics{}, nil
	} else {
		evaled, diags := e.Value(ectx)
		if diags.HasErrors() {
			return nil, diags, nil
		}

		res, err := noMetaJsonEncode(evaled)

		return res, hcl.Diagnostics{}, err
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

type ValidationBlock interface {
	HasValidationErrors() bool
	Encode() ([]byte, error)
}

// func (me *FileBlockEvaluation) HasValidationErrors() bool {
// 	return me.Validation != nil
// }
