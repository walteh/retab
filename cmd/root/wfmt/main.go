package wfmt

// `hclFmt` command recursively looks for hcl files in the directory tree starting at workingDir, and formats them
// based on the language style guides provided by Hashicorp. This is done using the official hcl2 library.

import (
	"context"
	"io"
	"reflect"

	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/walteh/retab/cmd/root/resolvers"
	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/externalwrite"
	"github.com/walteh/retab/pkg/format"
	"github.com/walteh/retab/pkg/hclwrite"
	"github.com/walteh/retab/pkg/protowrite"
	"github.com/walteh/snake"
	"github.com/walteh/terrors"
)

func Runner() snake.Runner {
	return snake.GenRunCommand_In05_Out01(&Handler{})
}

type Handler struct {
	Proto bool `usage:"format .proto files"`
	Dart  bool `usage:"format .dart files"`
	Tf    bool `usage:"format .tf files"`
	Hcl   bool `usage:"format .hcl files"`
}

func (me *Handler) Name() string {
	return "wfmt"
}

func (me *Handler) Description() string {
	return "format files with the hcl golang library, but with tabs"
}

func (me *Handler) Run(ctx context.Context, fls afero.Fs, fle afero.File, ecfg configuration.Provider, out snake.Stdout) error {
	fmtrs := []format.Provider{}

	if me.Hcl {
		fmtrs = append(fmtrs, hclwrite.NewFormatter())
	}

	if me.Proto {
		fmtrs = append(fmtrs, protowrite.NewFormatter())
	}

	if me.Dart {
		fmtrs = append(fmtrs, externalwrite.NewDartFormatter("dart"))
	}

	if me.Tf {
		fmtrs = append(fmtrs, externalwrite.NewTerraformFormatter("terraform"))
	}

	if len(fmtrs) == 0 {
		return terrors.New("no formatters specified")
	}

	flefmtrmap := map[format.Provider][]string{}

	for _, fmtr := range fmtrs {

		for _, s := range fmtr.Targets() {

			fles, err := resolvers.GetFileOrGlobDir(ctx, fls, fle, s)
			if err != nil {
				return err
			}

			zerolog.Ctx(ctx).Debug().Strs("files", fles).Str("fmtr", reflect.TypeOf(fmtr).String()).Msg("adding provider to map")

			for _, fle := range fles {

				flefmtrmap[fmtr] = append(flefmtrmap[fmtr], fle)
			}
		}

	}

	zerolog.Ctx(ctx).Debug().Any("flefmtrmap", flefmtrmap).Msg("flefmtrmap")

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
