package validate

import (
	"context"
	"fmt"

	"github.com/spf13/afero"
	"github.com/walteh/retab/pkg/hclread"
	"github.com/walteh/snake"
)

func Runner() snake.Runner {
	return snake.GenRunCommand_In03_Out01(&Handler{})
}

type Handler struct {
}

func (me *Handler) Name() string {
	return "validate"
}

func (me *Handler) Description() string {
	return "validate files defined in .retab files"
}

func (me *Handler) Run(ctx context.Context, fs afero.Fs, stdout snake.Stdout) error {
	// {*.retab.hcl}{.retab/*.retab}{.retab/*.retab.hcl}
	fles, err := afero.Glob(fs, "*.retab")
	if err != nil {
		return err
	}

	for _, fle := range fles {
		body, err := hclread.Process(ctx, fs, fle)
		if err != nil {
			return err
		}

		for _, blk := range body.File.Validation {
			fmt.Fprintf(stdout, "start[line=%d,col=%d] end[line=%d,col=%d] message[%s]\n", blk.Range.Start.Line, blk.Range.Start.Column, blk.Range.End.Line, blk.Range.End.Column, blk.Message)
		}
	}

	return nil
}
