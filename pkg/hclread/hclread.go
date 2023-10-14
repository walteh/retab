package hclread

import (
	"bytes"
	"context"
	"io"
	"path/filepath"

	"github.com/go-faster/errors"
	"github.com/spf13/afero"
)

func Process(ctx context.Context, fs afero.Fs, file string) ([]*BlockEvaluation, error) {
	opn, err := fs.Open(file)
	if err != nil {
		return nil, err
	}

	_, ectx, blks, err := NewEvaluation(ctx, opn)
	if err != nil {
		return nil, err
	}

	evals := make([]*BlockEvaluation, 0, len(blks.Blocks))

	for _, blk := range blks.Blocks {
		eval, err := NewBlockEvaluation(ctx, ectx, blk)
		if err != nil {
			return nil, err
		}

		evals = append(evals, eval)

	}

	return evals, nil
}

func (me *BlockEvaluation) WriteToFile(ctx context.Context, fs afero.Fs) error {
	out, erry := me.WriteToReader(ctx)
	if erry != nil {
		return errors.Wrapf(erry, "failed to encode block %q", me.Name)
	}

	if err := fs.MkdirAll(me.Dir, 0755); err != nil {
		return errors.Wrapf(err, "failed to create directory %q", me.Dir)
	}

	if err := afero.WriteReader(fs, filepath.Join(me.Dir, me.Name), out); err != nil {
		return errors.Wrapf(err, "failed to write file %q", me.Name)
	}

	return nil
}

func (me *BlockEvaluation) WriteToReader(ctx context.Context) (io.Reader, error) {
	out, erry := me.Encode()
	if erry != nil {
		return nil, errors.Wrapf(erry, "failed to encode block %q", me.Name)
	}

	return bytes.NewReader(out), nil
}
