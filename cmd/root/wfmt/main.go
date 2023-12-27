package wfmt

// `hclFmt` command recursively looks for hcl files in the directory tree starting at workingDir, and formats them
// based on the language style guides provided by Hashicorp. This is done using the official hcl2 library.

import (
	"context"
	"io"

	"github.com/spf13/afero"
	"github.com/walteh/retab/cmd/root/resolvers"
	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/externalwrite"
	"github.com/walteh/retab/pkg/format"
	"github.com/walteh/retab/pkg/hclwrite"
	"github.com/walteh/retab/pkg/protowrite"
	"github.com/walteh/snake"
)

func Runner() snake.Runner {
	return snake.GenRunCommand_In05_Out01(&Handler{})
}

type Handler struct {
}

func (me *Handler) Name() string {
	return "wfmt"
}

func (me *Handler) Description() string {
	return "format files with the hcl golang library, but with tabs"
}

func (me *Handler) Run(ctx context.Context, fls afero.Fs, fle afero.File, ecfg configuration.Provider, out snake.Stdout) error {
	fmtrs := []format.Provider{}

	fmtrs = append(fmtrs, hclwrite.NewFormatter())
	fmtrs = append(fmtrs, protowrite.NewFormatter())
	fmtrs = append(fmtrs, externalwrite.NewDartFormatter("dart"))
	fmtrs = append(fmtrs, externalwrite.NewTerraformFormatter("terraform"))

	flefmtrmap := map[format.Provider][]string{}

	for _, fmtr := range fmtrs {

		for _, s := range fmtr.Targets() {

			fles, err := resolvers.GetFileOrGlobDir(ctx, fls, fle, s)
			if err != nil {
				return err
			}

			for _, fle := range fles {
				flefmtrmap[fmtr] = append(flefmtrmap[fmtr], fle)
			}
		}

	}

	for fmtr, fles := range flefmtrmap {
		err := resolvers.ForAllFilesAtSameTime(ctx, fls, fles, func(ctx context.Context, fle afero.File) (io.Reader, error) {
			return format.Format(ctx, fmtr, ecfg, fle.Name(), fle)
		})
		if err != nil {
			return err
		}
	}

	return nil
}
