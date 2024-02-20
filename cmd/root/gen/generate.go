package gen

import (
	"context"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/walteh/retab/cmd/root/resolvers"
	"github.com/walteh/retab/pkg/hclread"
	"github.com/walteh/snake"
)

type Handler struct {
}

func (me *Handler) RegisterRunFunc() snake.RunFunc {
	return snake.GenRunCommand_In03_Out01(me)
}

func (me *Handler) CobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen",
		Short: "generate files defined in .retab files",
	}

	return cmd
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
