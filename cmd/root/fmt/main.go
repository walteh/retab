package fmt

// `hclFmt` command recursively looks for hcl files in the directory tree starting at workingDir, and formats them
// based on the language style guides provided by Hashicorp. This is done using the official hcl2 library.

import (
	"context"
	"errors"

	"github.com/mattn/go-zglob"
	"github.com/spf13/afero"
	"github.com/walteh/retab/pkg/bufwrite"
	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/externalwrite"
	"github.com/walteh/retab/pkg/format"
	"github.com/walteh/retab/pkg/hclwrite"
	"github.com/walteh/snake"
)

func Runner() snake.Runner {
	return snake.GenRunCommand_In04_Out01(&Handler{})
}

type Handler struct {
}

func (me *Handler) Name() string {
	return "fmt"
}

func (me *Handler) Description() string {
	return "format files with the official buf library, but with tabs"
}

func (me *Handler) Run(ctx context.Context, fs afero.Fs, fle afero.File, ecfg configuration.Provider) error {

	fmtrs := []format.Provider{
		hclwrite.NewHclFormatter(),
		bufwrite.NewBufFormatter(),
		externalwrite.NewDartFormatter("dart"),
	}

	for _, fmtr := range fmtrs {
		for _, target := range fmtr.Targets() {
			// targets are glob patterns
			matches, err := zglob.Glob(target)
			if err != nil {
				return err
			}

			if len(matches) == 0 {
				continue
			}

			return format.Format(ctx, fmtr, ecfg, fs, fle)
		}
	}

	return errors.New("no targets found")
}
