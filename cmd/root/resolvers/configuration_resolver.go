package resolvers

import (
	"context"

	"github.com/spf13/afero"
	"github.com/walteh/retab/pkg/format"
	"github.com/walteh/retab/pkg/format/editorconfig"
	"github.com/walteh/snake"
	"github.com/walteh/terrors"
)

func ConfigurationRunner() snake.Runner {
	return snake.GenRunResolver_In02_Out02(&ConfigurationResolver{})
}

type ConfigurationResolver struct {
	UseTabs                bool `default:"true" help:"Use tabs instead of spaces"`
	IndentSize             int  `default:"4" help:"Number of spaces or tabs to use for indentation"`
	TrimMultipleEmptyLines bool `default:"true" help:"Trim multiple empty lines"`
}

func (me *ConfigurationResolver) Run(ctx context.Context, fls afero.Fs) (format.ConfigurationProvider, error) {

	efg, err := editorconfig.NewEditorConfigConfigurationProvider(ctx, fls)
	if err != nil {
		return nil, terrors.Wrap(err, "failed to get editorconfig definition")
	}

	return efg, nil
}
