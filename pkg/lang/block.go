package lang

import (
	"context"
	"encoding/json"
	"regexp"
	"slices"

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
}

type FileBlockEvaluation struct {
	Name          string
	Schema        string
	Path          string
	OrderedOutput yaml.MapSlice
	RawOutput     any
	Source        string
}

func evalGenBlock(ctx context.Context, sctx *SudoContext, file *BodyBuilder) (res *FileBlockEvaluation, diags hcl.Diagnostics, err error) {

	gblk := sctx.Meta.(*GenBlockMeta)
	fblock := gblk.HCL

	ectx := sctx.BuildStaticEvalContextWithFileData(fblock.TypeRange.Filename)

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

	var d hclsyntax.Expression

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

		switch attr.Name {
		case "path":
			unrk, _ := sctx.Map["path"].ToValue().Unmark()

			blk.Path = unrk.AsString()

			blk.Path = sanatizeGenPath(blk.Path)

			ectx.Functions["ref"] = NewRefFunctionFromPath(ctx, blk.Path)

			defer delete(ectx.Functions, "ref")
		case "schema":
			unrk, _ := sctx.Map["schema"].ToValue().Unmark()

			blk.Schema = unrk.AsString()
		case "data":

			cnt := yaml.MapSlice{}

			// sctx.Map["data"].ToValue()

			dat := sctx.Map["data"].ToValue()

			slc, err := UnmarkToSortedArray(dat)
			if err != nil {
				return nil, hcl.Diagnostics{}, terrors.Wrap(err, "problem encoding yaml")
			}

			if x, ok := slc.(yaml.MapSlice); ok {
				cnt = append(cnt, x...)
			}
			if x, ok := slc.(yaml.MapItem); ok {
				cnt = append(cnt, x)
			}

			mar, err := json.Marshal(cnt)
			if err != nil {
				return nil, hcl.Diagnostics{}, terrors.Wrap(err, "problem encoding json")
			}

			var unmar any
			err = json.Unmarshal(mar, &unmar)
			if err != nil {
				return nil, hcl.Diagnostics{}, terrors.Wrap(err, "problem encoding json")
			}

			blk.OrderedOutput = cnt

			blk.RawOutput = unmar

			d = attr.Expr

		default:
			continue
		}
	}

	if d == nil {
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
			if lerr, err := LoadValidationErrors(ctx, d, ectx, errv, file, sctx.Map["data"]); err != nil {
				return nil, hcl.Diagnostics{}, terrors.Wrap(err, "problem loading validation errors")
			} else {
				diags = append(diags, lerr...)
			}
		}
	}

	return blk, diags, nil

}

func NewGenBlockEvaluation(ctx context.Context, sctx *SudoContext, file *BodyBuilder) (res []*FileBlockEvaluation, diags hcl.Diagnostics, err error) {

	fblocks := sctx.GetAllFileLevelBlocksOfType("gen")

	if len(fblocks) == 0 {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "missing gen block",
			Detail:   "a file must have at least one gen block",
			// Subject:  &file.Blocks[0].TypeRange,
		}}, nil
	}

	output := make([]*FileBlockEvaluation, 0)

	for _, fblock := range fblocks {
		res, diags, err := evalGenBlock(ctx, fblock, file)
		if err != nil || diags.HasErrors() {
			return nil, diags, err
		}

		output = append(output, res)
	}

	return output, diags, nil

}

func EncodeExpression(e hclsyntax.Expression, ectx *hcl.EvalContext) (any, hcl.Diagnostics, error) {

	if x, ok := e.(*hclsyntax.ObjectConsExpr); ok {

		type kv struct {
			index int
			key   string
			value any
		}
		group := make([]kv, 0)
		for _, rr := range x.Items {
			kvf, diags := rr.KeyExpr.Value(ectx)
			if diags.HasErrors() {
				return nil, diags, nil
			}
			if kvf.AsString() == MetaKey {
				continue
			}
			rz, diags, errd := EncodeExpression(rr.ValueExpr, ectx)
			if errd != nil || diags.HasErrors() {
				return nil, diags, errd
			}
			if rz == nil {
				continue
			}

			group = append(group, kv{index: rr.KeyExpr.Range().Start.Line, key: kvf.AsString(), value: rz})
		}

		slices.SortFunc(group, func(a, b kv) int {
			return a.index - b.index
		})

		wrk := make(yaml.MapSlice, 0)
		for _, g := range group {
			wrk = append(wrk, yaml.MapItem{Key: g.key, Value: g.value})
		}

		return wrk, hcl.Diagnostics{}, nil
	} else if x, ok := e.(*hclsyntax.TupleConsExpr); ok {
		wrk := make([]any, 0)
		for _, exp := range x.Exprs {
			r, diags, err := EncodeExpression(exp, ectx)
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
	ok, _ = ok.Unmark()
	oks := reg1.ReplaceAllString(ok.AsString(), "")
	oks = reg2.ReplaceAllString(oks, "{")
	ok = cty.StringVal(oks)

	var ok2 interface{}
	err = json.Unmarshal([]byte(ok.AsString()), &ok2)
	if err != nil {
		return nil, err
	}
	return ok2, nil
}

type ValidationBlock interface {
	HasValidationErrors() bool
	Encode() ([]byte, error)
}
