package dart

import (
	"context"

	"github.com/spf13/afero"
	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/externalwrite"
	"github.com/walteh/retab/pkg/format"
	"github.com/walteh/snake"
)

func Runner() snake.Runner {
	return snake.GenRunCommand_In04_Out01(&Handler{})
}

type Handler struct {
}

func (me *Handler) Name() string {
	return "dart"
}

func (me *Handler) Description() string {
	return "format dart files with your local version of dart, but with tabs"
}

func (me *Handler) Run(ctx context.Context, fs afero.Fs, fle afero.File, ecfg configuration.Provider) error {
	fourmatter := externalwrite.NewDartFormatter("dart")
	return format.Format(ctx, fourmatter, ecfg, fs, fle)
}
