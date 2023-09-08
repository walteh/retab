package buf

// `hclFmt` command recursively looks for hcl files in the directory tree starting at workingDir, and formats them
// based on the language style guides provided by Hashicorp. This is done using the official hcl2 library.

import (
	"context"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/walteh/retab/pkg/bufwrite"
	"github.com/walteh/retab/pkg/format"
	"github.com/walteh/snake"
)

var _ snake.Snakeable = (*Handler)(nil)

type Handler struct {
	File       string `arg:"" default:" " name:"file" help:"The hcl file to format."`
	WorkingDir string `name:"working-dir" help:"The working directory to use. Defaults to the current directory."`
}

func (me *Handler) BuildCommand(_ context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buf",
		Short: "format proto files with the official buf library, but with tabs",
	}
	cmd.Args = cobra.ExactArgs(1)

	cmd.Flags().StringVar(&me.WorkingDir, "working-dir", "", "The working directory to use. Defaults to the current directory.")

	return cmd
}

func (me *Handler) ParseArguments(_ context.Context, _ *cobra.Command, file []string) error {

	me.File = file[0]

	return nil

}

func (me *Handler) Run(ctx context.Context, fs afero.Fs) error {

	fmtr := bufwrite.NewBufFormatter()

	return format.Format(ctx, fmtr, fs, me.File, me.WorkingDir)
}
