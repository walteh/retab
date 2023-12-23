package hclread

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"path/filepath"
	"strings"

	"github.com/go-faster/errors"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

func Process(ctx context.Context, fs afero.Fs, file string) (*FullEvaluation, error) {
	opn, err := fs.Open(file)
	if err != nil {
		return nil, err
	}

	_, ectx, blks, err := NewEvaluation(ctx, opn)
	if err != nil {
		return nil, err
	}

	eval, err := NewFullEvaluation(ctx, ectx, blks)
	if err != nil {
		return nil, err
	}

	return eval, nil
}

func (me *FullEvaluation) WriteToFile(ctx context.Context, fs afero.Fs) error {
	out, erry := me.WriteToReader(ctx)
	if erry != nil {
		return errors.Wrapf(erry, "failed to encode block %q", me.File.Name)
	}

	if err := fs.MkdirAll(me.File.Dir, 0755); err != nil {
		return errors.Wrapf(err, "failed to create directory %q", me.File.Dir)
	}

	if err := afero.WriteReader(fs, filepath.Join(me.File.Dir, me.File.Name), out); err != nil {
		return errors.Wrapf(err, "failed to write file %q", me.File.Name)
	}

	return nil
}

func (me *FullEvaluation) WriteToReader(_ context.Context) (io.Reader, error) {
	out, erry := me.Encode()
	if erry != nil {
		return nil, errors.Wrapf(erry, "failed to encode block %q", me.File.Name)
	}

	return bytes.NewReader(out), nil
}

func (me *FullEvaluation) Encode() ([]byte, error) {
	arr := strings.Split(me.File.Name, ".")
	if len(arr) < 2 {
		return nil, errors.Errorf("invalid file name [%s] - missing extension", me.File.Name)
	}

	content := me.File.Content
	for _, blk := range me.Other {
		content[blk.Name] = blk.Content
	}

	switch arr[len(arr)-1] {
	case "json":
		return json.MarshalIndent(content, "", "\t")
	case "yaml":
		buf := bytes.NewBuffer(nil)
		enc := yaml.NewEncoder(buf)
		enc.SetIndent(4)
		err := enc.Encode(content)
		if err != nil {
			return nil, err
		}

		strWithTabsRemovedFromHeredoc := strings.ReplaceAll(buf.String(), "\t", "")

		return []byte(strWithTabsRemovedFromHeredoc), nil

	default:
		return nil, errors.Errorf("unknown file extension [%s] in %s", arr[len(arr)-1], me.File.Name)
	}
}
