package hcl

// `hclFmt` command recursively looks for hcl files in the directory tree starting at workingDir, and formats them
// based on the language style guides provided by Hashicorp. This is done using the official hcl2 library.

import (
	"context"
	"errors"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/walteh/retab/pkg/format"
	"github.com/walteh/retab/pkg/hclwrite"
	"github.com/walteh/snake"
)

type Handler struct {
	WorkingDir string
	File       string
}

var _ snake.Cobrad = (*Handler)(nil)

func (me *Handler) Cobra() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hcl [file]",
		Short: "format hcl files with the official hcl2 library, but with tabs",
	}

	cmd.Args = func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("requires a file argument")
		}
		me.File = args[0]
		if me.File == "" {
			return errors.New("no file provided")
		}
		return nil
	}

	cmd.Flags().StringVar(&me.WorkingDir, "working-dir", "", "The working directory to use. Defaults to the current directory.")

	return cmd
}

func (me *Handler) Run(ctx context.Context, fs afero.Fs) error {

	fourmatter := hclwrite.NewHclFormatter()

	return format.Format(ctx, fourmatter, fs, me.File, me.WorkingDir)
}
