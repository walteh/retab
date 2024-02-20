package validate

import (
	"context"
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/walteh/retab/cmd/root/resolvers"
	"github.com/walteh/retab/pkg/hclread"
	"github.com/walteh/snake"
)

type Handler struct {
}

func (me *Handler) RegisterRunFunc() snake.RunFunc {
	return snake.GenRunCommand_In04_Out01(me)
}

func (me *Handler) CobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "validate files defined in .retab files",
	}

	return cmd
}

func (me *Handler) Run(ctx context.Context, fls afero.Fs, fle afero.File, stdout snake.Stdout) error {

	fles, err := resolvers.GetFileOrGlobDir(ctx, fls, fle, ".retab/*.retab")
	if err != nil {
		return err
	}

	for _, fle := range fles {
		_, diags, err := hclread.Process(ctx, fls, fle)
		if err != nil {
			return err
		}

		for _, blk := range diags {
			fmt.Fprintf(stdout, "%+v\n", blk)
			// fmt.Fprintf(stdout, "start[line=%d,col=%d] end[line=%d,col=%d] message[%s]\n",
			// 	blk.Expression.Range().Start.Line,
			// 	blk.Expression.Range().Start.Column,
			// 	blk.Expression.Range().End.Line,
			// 	blk.Expression.Range().End.Column,
			// 	blk.Detail,
			// )
		}
	}

	return nil
}
