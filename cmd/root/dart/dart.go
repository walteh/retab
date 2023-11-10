package dart

import (
	"context"
	"errors"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/externalwrite"
	"github.com/walteh/retab/pkg/format"
	"github.com/walteh/snake"
)

type Handler struct {
	WorkingDir string
	File       string
}

var _ snake.Cobrad = (*Handler)(nil)

func (me *Handler) Cobra() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dart <file>",
		Short: "format dart files with your local version of dart, but with tabs",
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

func (me *Handler) Run(ctx context.Context, fs afero.Fs, ecfg configuration.Provider) error {

	fourmatter := externalwrite.NewDartFormatter("dart")

	return format.Format(ctx, fourmatter, ecfg, fs, me.File, me.WorkingDir)
}
