package generate

import (
	"context"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/walteh/retab/pkg/hclread"
	"github.com/walteh/snake"
)

var _ snake.Cobrad = (*Handler)(nil)

type Handler struct {
	File string
}

func (me *Handler) Cobra() *cobra.Command {
	cmd := &cobra.Command{
		Use: "generate",
	}

	cmd.Args = cobra.ExactArgs(0)

	cmd.Flags().StringVar(&me.File, "file", "retab.hcl", "location of the retab.hcl file")

	return cmd
}

func (me *Handler) Run(ctx context.Context, fs afero.Fs) error {

	body, err := hclread.Process(ctx, fs, me.File)
	if err != nil {
		return err
	}
	err = body.WriteToFile(ctx, fs)
	if err != nil {
		return err
	}

	return nil
}
