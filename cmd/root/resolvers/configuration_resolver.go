package resolvers

import (
	"context"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/editorconfig"
	"github.com/walteh/snake"
)

var _ snake.Flagged = (*ConfigurationResolver)(nil)

type ConfigurationResolver struct {
	useTabs                bool
	indentSize             int
	trimMultipleEmptyLines bool
}

func (me *ConfigurationResolver) Flags(flgs *pflag.FlagSet) {
	flgs.BoolVar(&me.useTabs, "use-tabs", true, "Use tabs instead of spaces")
	flgs.IntVar(&me.indentSize, "indent-size", 4, "Number of spaces or tabs to use for indentation")
	flgs.BoolVar(&me.trimMultipleEmptyLines, "trim-multiple-empty-lines", true, "Trim multiple empty lines")
}

func (me *ConfigurationResolver) Run(ctx context.Context, cmd *cobra.Command, fle afero.File) (configuration.Provider, error) {

	efg, err := editorconfig.NewEditorConfigConfigurationProvider(ctx, fle.Name())
	if err != nil {
		cmd.Println("No .editorconfig file found, using default configuration")
		return configuration.NewBasicConfigurationProvider(me.useTabs, me.indentSize, me.trimMultipleEmptyLines), nil
	}

	if !cmd.Flags().Lookup("use-tabs").Changed {
		me.useTabs = efg.UseTabs()
	}

	if cmd.Flags().Lookup("indent-size").Changed {
		me.indentSize = efg.IndentSize()
	}

	if cmd.Flags().Lookup("trim-multiple-empty-lines").Changed {
		me.trimMultipleEmptyLines = efg.TrimMultipleEmptyLines()
	}

	return configuration.NewBasicConfigurationProvider(me.useTabs, me.indentSize, me.trimMultipleEmptyLines), nil
}
