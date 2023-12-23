package dart

import (
	"context"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/externalwrite"
	"github.com/walteh/retab/pkg/format"
	"github.com/walteh/snake"
)

type Handler struct {
}

var _ snake.Cobrad = (*Handler)(nil)

func (me *Handler) Cobra() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dart <file>",
		Short: "format dart files with your local version of dart, but with tabs",
	}

	return cmd
}

func (me *Handler) Run(ctx context.Context, fs afero.Fs, fle afero.File, ecfg configuration.Provider) error {

	fourmatter := externalwrite.NewDartFormatter("dart")

	return format.Format(ctx, fourmatter, ecfg, fs, fle)
}
