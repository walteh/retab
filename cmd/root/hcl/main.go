package hcl

// `hclFmt` command recursively looks for hcl files in the directory tree starting at workingDir, and formats them
// based on the language style guides provided by Hashicorp. This is done using the official hcl2 library.

import (
	"context"

	"github.com/spf13/afero"
	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/format"
	"github.com/walteh/retab/pkg/hclwrite"
	"github.com/walteh/snake"
)

func Runner() snake.Runner {
	return snake.GenRunCommand_In04_Out01(&Handler{})
}

type Handler struct {
}

// name
func (me *Handler) Name() string {
	return "hcl"
}

// description
func (me *Handler) Description() string {
	return "format hcl files with the official hcl2 library, but with tabs"
}

func (me *Handler) Run(ctx context.Context, fs afero.Fs, fle afero.File, ecfg configuration.Provider) error {
	fmtr := hclwrite.NewFormatter()
	return format.Format(ctx, fmtr, ecfg, fs, fle)
}
