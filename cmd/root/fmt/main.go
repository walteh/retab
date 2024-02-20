package fmt

// `hclFmt` command recursively looks for hcl files in the directory tree starting at workingDir, and formats them
// based on the language style guides provided by Hashicorp. This is done using the official hcl2 library.

import (
	"context"
	"io"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/walteh/retab/cmd/root/resolvers"
	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/format"
	"github.com/walteh/retab/pkg/hclwrite"
	"github.com/walteh/snake"
)

type Handler struct {
	All bool `usage:"format all supported files, not just .retab files - .hcl, .proto, .tf, .tfvars, .dart"`
}

func (me *Handler) RegisterRunFunc() snake.RunFunc {
	return snake.GenRunCommand_In05_Out01(me)
}

func (me *Handler) CobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fmt",
		Short: "format .retab files with the hcl golang library, but with tabs - also format other files if --all is specified",
	}

	return cmd
}

func (me *Handler) Run(ctx context.Context, fls afero.Fs, fle afero.File, ecfg configuration.Provider, out snake.Stdout) error {

	fles, err := resolvers.GetFileOrGlobDir(ctx, fls, fle, ".retab/*.retab")
	if err != nil {
		return err
	}

	fmtr := hclwrite.NewFormatter()

	err = resolvers.ForAllFilesAtSameTime(ctx, fls, fles, func(ctx context.Context, fle afero.File) (io.Reader, error) {
		return format.Format(ctx, fmtr, ecfg, fle.Name(), fle)
	})

	if err != nil {
		return err
	}

	return nil
}
