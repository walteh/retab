package resolvers

import (
	"context"
	"fmt"

	"github.com/spf13/afero"
	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/editorconfig"
	"github.com/walteh/snake"
)

func ConfigurationRunner() snake.Runner {
	return snake.GenRunResolver_In03_Out02(&ConfigurationResolver{})
}

type ConfigurationResolver struct {
	UseTabs                bool `default:"true" help:"Use tabs instead of spaces"`
	IndentSize             int  `default:"4" help:"Number of spaces or tabs to use for indentation"`
	TrimMultipleEmptyLines bool `default:"true" help:"Trim multiple empty lines"`
}

func (me *ConfigurationResolver) Run(ctx context.Context, fle afero.File, out snake.Stdout) (configuration.Provider, error) {

	efg, err := editorconfig.NewEditorConfigConfigurationProvider(ctx, fle.Name())
	if err != nil {
		fmt.Fprintln(out, "No .editorconfig file found, using default configuration")
		return configuration.NewBasicConfigurationProvider(me.UseTabs, me.IndentSize, me.TrimMultipleEmptyLines), nil
	}

	return efg, nil
}
