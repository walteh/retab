package wfmt

// `hclFmt` command recursively looks for hcl files in the directory tree starting at workingDir, and formats them
// based on the language style guides provided by Hashicorp. This is done using the official hcl2 library.

import (
	"context"
	"io"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/walteh/retab/cmd/root/resolvers"
	"github.com/walteh/retab/pkg/format"
	"github.com/walteh/retab/pkg/format/cmdfmt"
	"github.com/walteh/retab/pkg/format/hclfmt"
	"github.com/walteh/retab/pkg/format/protofmt"
	"github.com/walteh/snake"
	"github.com/walteh/terrors"
)

type Handler struct {
	Proto bool `usage:"format .proto files"`
	Dart  bool `usage:"format .dart files"`
	Tf    bool `usage:"format .tf files"`
	Hcl   bool `usage:"format .hcl files"`
}

func (me *Handler) RegisterRunFunc() snake.RunFunc {
	return snake.GenRunCommand_In05_Out01(me)
}

func (me *Handler) CobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wfmt",
		Short: "format files with the hcl golang library, but with tabs",
	}

	return cmd
}

func (me *Handler) Run(ctx context.Context, fls afero.Fs, fle afero.File, ecfg format.ConfigurationProvider, out snake.Stdout) error {
	fmtrs := []format.Provider{}

	if me.Hcl {
		fmtrs = append(fmtrs, hclfmt.NewFormatter())
	}

	if me.Proto {
		fmtrs = append(fmtrs, protofmt.NewFormatter())
	}

	if me.Dart {
		fmtrs = append(fmtrs, cmdfmt.NewDartFormatter("dart"))
	}

	if me.Tf {
		fmtrs = append(fmtrs, cmdfmt.NewTerraformFormatter("terraform"))
	}

	if len(fmtrs) == 0 {
		return terrors.New("no formatters specified")
	}

	if len(fmtrs) > 1 {
		return terrors.New("only one formatter can be specified")
	}

	fles, err := resolvers.GetFileOrGlobDir(ctx, fls, fle, "*")
	if err != nil {
		return err
	}

	err = resolvers.ForAllFilesAtSameTime(ctx, fls, fles, func(ctx context.Context, fle afero.File) (io.Reader, error) {
		return format.Format(ctx, fmtrs[0], ecfg, fle.Name(), fle)
	})
	if err != nil {
		return err
	}

	return nil
}
