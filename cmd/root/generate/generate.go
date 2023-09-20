package generate

import (
	"context"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/walteh/retab/pkg/hclread"
	"github.com/walteh/snake"
)

var _ snake.Snakeable = (*Handler)(nil)

type Handler struct {
	File string
}

func (me *Handler) BuildCommand(_ context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use: "generate",
	}

	cmd.Args = cobra.ExactArgs(0)

	cmd.Flags().StringVar(&me.File, "file", "retab.hcl", "location of the retab.hcl file")

	return cmd
}

func (me *Handler) ParseArguments(_ context.Context, _ *cobra.Command, file []string) error {

	return nil

}

func (me *Handler) Run(ctx context.Context, fs afero.Fs, cmd *cobra.Command) error {

	body, err := hclread.Process(ctx, fs, me.File)
	if err != nil {
		return err
	}

	for _, blk := range body {
		if blk.Validation != nil {
			return blk.Validation
		}

		err := blk.WriteToFile(ctx, fs)
		if err != nil {
			return err
		}
	}

	return nil
}