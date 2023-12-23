package hcl

// `hclFmt` command recursively looks for hcl files in the directory tree starting at workingDir, and formats them
// based on the language style guides provided by Hashicorp. This is done using the official hcl2 library.

import (
	"context"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/format"
	"github.com/walteh/retab/pkg/hclwrite"
	"github.com/walteh/snake"
)

type Handler struct {
}

var _ snake.Cobrad = (*Handler)(nil)

func (me *Handler) Cobra() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hcl",
		Short: "format hcl files with the official hcl2 library, but with tabs",
	}

	return cmd
}

func (me *Handler) Run(ctx context.Context, fs afero.Fs, fle afero.File, ecfg configuration.Provider) error {

	fmtr := hclwrite.NewHclFormatter()

	return format.Format(ctx, fmtr, ecfg, fs, fle)
}
