package proto

import (
	"context"

	"github.com/spf13/afero"
	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/format"
	"github.com/walteh/retab/pkg/protowrite"
	"github.com/walteh/snake"
)

func Runner() snake.Runner {
	return snake.GenRunCommand_In04_Out01(&Handler{})
}

type Handler struct {
}

func (me *Handler) Name() string {
	return "proto"
}

func (me *Handler) Description() string {
	return "format proto files with the official buf library, but with tabs"
}

func (me *Handler) Run(ctx context.Context, fs afero.Fs, fle afero.File, ecfg configuration.Provider) error {
	fmtr := protowrite.NewFormatter()
	return format.Format(ctx, fmtr, ecfg, fs, fle)
}
