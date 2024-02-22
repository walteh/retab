package gen

import (
	"context"
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/walteh/retab/cmd/root/resolvers"
	"github.com/walteh/retab/pkg/lang"
	"github.com/walteh/snake"
)

type Handler struct {
	Verbose bool `name:"verbose" usage:"verbose output" default:"false"`
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

	bodies, diags, err := lang.ProccessBulk(ctx, fls, fles)
	if err != nil {
		return err
	}

	if diags.HasErrors() {
		if me.Verbose {
			for _, diag := range diags {
				fmt.Println(diag)
			}
		}
		return diags
	}

	for _, body := range bodies {

		err = body.WriteToFile(ctx, fls)
		if err != nil {
			return err
		}
	}

	return nil
}
