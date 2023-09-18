package yaml

import (
	"context"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/walteh/retab/pkg/format"
	"github.com/walteh/retab/pkg/yamlwrite"
	"github.com/walteh/snake"
)

var _ snake.Snakeable = (*Handler)(nil)

type Handler struct {
	File       string `arg:"" default:" " name:"file" help:"The hcl file to format."`
	WorkingDir string `name:"working-dir" help:"The working directory to use. Defaults to the current directory."`
}

func (me *Handler) BuildCommand(_ context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fmt",
		Short: "format hcl files with the official hcl2 library, but with tabs",
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

	fourmatter := yamlwrite.NewYamlFormatter()

	// decode from yaml into json

	return format.Format(ctx, fourmatter, fs, me.File, me.WorkingDir)
}
