package gen

import (
	"context"

	"github.com/spf13/afero"
	"github.com/walteh/retab/pkg/hclread"
	"github.com/walteh/snake"
)

func Runner() snake.Runner {
	return snake.GenRunCommand_In02_Out01(&Handler{})
}

type Handler struct {
}

func (me *Handler) Name() string {
	return "gen"
}

func (me *Handler) Description() string {
	return "generate files defined in .retab files"
}

func (me *Handler) Run(ctx context.Context, fs afero.Fs) error {
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
		err = body.WriteToFile(ctx, fs)
		if err != nil {
			return err
		}
	}

	return nil
}
