package gen

import (
	"context"

	"github.com/spf13/afero"
	"github.com/walteh/retab/cmd/root/resolvers"
	"github.com/walteh/retab/pkg/hclread"
	"github.com/walteh/snake"
)

func Runner() snake.Runner {
	return snake.GenRunCommand_In03_Out01(&Handler{})
}

type Handler struct {
}

func (me *Handler) Name() string {
	return "gen"
}

func (me *Handler) Description() string {
	return "generate files defined in .retab files"
}

func (me *Handler) Run(ctx context.Context, fls afero.Fs, fle afero.File) error {

	fles, err := resolvers.GetFileOrGlobDir(ctx, fls, fle, ".retab/*.retab")
	if err != nil {
		return err
	}

	for _, fle := range fles {
		body, diags, err := hclread.Process(ctx, fls, fle)
		if err != nil {
			return err
		}

		if diags.HasErrors() {
			return diags
		}

		err = body.WriteToFile(ctx, fls)
		if err != nil {
			return err
		}
	}

	return nil
}
