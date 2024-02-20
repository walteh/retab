package hclread

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/afero"
	"github.com/walteh/terrors"
	"github.com/walteh/yaml"
)

func ProccessBulk(ctx context.Context, fs afero.Fs, files []string) ([]*FileBlockEvaluation, hcl.Diagnostics, error) {
	var out []*FileBlockEvaluation
	diags := hcl.Diagnostics{}

	globalFiles := make(map[string]cty.Value)

	global := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
	}

	ectxs := make(map[string]*hcl.EvalContext)
	blks := make(map[string]*hclsyntax.Body)

	for _, file := range files {

		opn, err := afero.ReadFile(fs, file)
		if err != nil {
			return nil, nil, err
		}

		_, ectx, blk, diags, err := NewContextFromFile(ctx, opn, file)
		if err != nil || diags.HasErrors() {
			return nil, diags, err
		}

		ectxs[file] = ectx
		blks[file] = blk

		basenoext := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))

		globalFiles[basenoext] = cty.ObjectVal(ectx.Variables)

	}

	global.Variables[FilesKey] = cty.ObjectVal(globalFiles)

	gfuncs := NewGlobalContextualizedFunctionMap(global)

	for _, file := range files {

		for k, v := range gfuncs {
			ectxs[file].Functions[k] = v
		}

		eval, diags, err := NewGenBlockEvaluation(ctx, ectxs[file], blks[file])
		if err != nil || diags.HasErrors() {
			return nil, diags, err
		}

		out = append(out, eval)
	}

	return out, diags, nil
}

func Process(ctx context.Context, fs afero.Fs, file string) (*FileBlockEvaluation, hcl.Diagnostics, error) {
	opn, err := afero.ReadFile(fs, file)
	if err != nil {
		return nil, nil, err
	}

	_, ectx, blks, diags, err := NewContextFromFile(ctx, opn, file)
	if err != nil || diags.HasErrors() {
		return nil, diags, err
	}

	eval, diags, err := NewGenBlockEvaluation(ctx, ectx, blks)
	if err != nil || diags.HasErrors() {
		return nil, diags, err
	}

	return eval, diags, nil
}

func (me *FileBlockEvaluation) WriteToFile(ctx context.Context, fs afero.Fs) error {
	out, erry := me.WriteToReader(ctx)
	if erry != nil {
		return terrors.Wrapf(erry, "failed to encode block %q", me.Name)
	}

	if err := fs.MkdirAll(filepath.Dir(me.Path), 0755); err != nil {
		return terrors.Wrapf(err, "failed to create directory %q", me.Path)
	}

	if err := afero.WriteReader(fs, me.Path, out); err != nil {
		return terrors.Wrapf(err, "failed to write file %q", me.Name)
	}

	return nil
}

func (me *FileBlockEvaluation) WriteToReader(ctx context.Context) (io.Reader, error) {
	out, erry := me.Encode()
	if erry != nil {
		return nil, terrors.Wrapf(erry, "failed to encode block %q", me.Name)
	}

	return bytes.NewReader(out), nil
}

func (me *FileBlockEvaluation) Encode() ([]byte, error) {

	arr := strings.Split(me.Path, ".")
	if len(arr) < 2 {
		return nil, terrors.Errorf("invalid file name [%s] - missing extension", me.Name)
	}

	header := fmt.Sprintf(`# code generated by retab. DO NOT EDIT.
# join the fight against yaml @ github.com/walteh/retab

# source: %q

`, me.Source)

	switch arr[len(arr)-1] {
	case "jsonc", "code-workspace":

		buf := bytes.NewBuffer(nil)
		enc := json.NewEncoder(buf)
		enc.SetIndent("", "\t")

		err := enc.Encode(me.OrderedOutput)
		if err != nil {
			return nil, err
		}

		return []byte(strings.ReplaceAll(header, "#", "//") + buf.String()), nil
	case "json":
		return json.MarshalIndent(me.OrderedOutput, "", "\t")
	case "yaml", "yml":
		buf := bytes.NewBuffer(nil)
		enc := yaml.NewEncoder(buf)
		// enc.SetIndent(4)
		defer enc.Close()

		err := enc.Encode(me.OrderedOutput)
		if err != nil {
			return nil, err
		}

		strWithTabsRemovedFromHeredoc := strings.ReplaceAll(buf.String(), "\t", "")

		return []byte(header + strWithTabsRemovedFromHeredoc), nil

	default:
		return nil, terrors.Errorf("unknown file extension [%s] in %s", arr[len(arr)-1], me.Name)
	}
}
