package gen

import (
	"context"

	"github.com/spf13/afero"
	"github.com/walteh/retab/pkg/hclread"
	"github.com/walteh/snake"
)

func Runner() snake.Runner {
	return snake.GenRunCommand_In02_Out01(&Handler{})
}

type Handler struct {
	File string `default:"retab.hcl" help:"The file to read the configuration from."`
}

func (me *Handler) Name() string {
	return "gen"
}

func (me *Handler) Description() string {
	return "generate files defined in retab.hcl"
}

func (me *Handler) Run(ctx context.Context, fs afero.Fs) error {

	body, err := hclread.Process(ctx, fs, me.File)
	if err != nil {
		return err
	}
	err = body.WriteToFile(ctx, fs)
	if err != nil {
		return err
	}

	return nil
}
