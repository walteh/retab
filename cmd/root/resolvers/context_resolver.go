package resolvers

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/walteh/retab/version"
	"github.com/walteh/snake"
	szerolog "github.com/walteh/snake/zerolog"
)

var _ snake.Flagged = (*ContextResolver)(nil)

type ContextResolver struct {
	Quiet   bool
	Debug   bool
	Version bool
}

func (me *ContextResolver) Flags(flgs *pflag.FlagSet) {
	flgs.BoolVarP(&me.Quiet, "quiet", "q", false, "Do not print any output")
	flgs.BoolVarP(&me.Debug, "debug", "d", false, "Print debug output")
	flgs.BoolVarP(&me.Version, "version", "v", false, "Print version and exit")
}

func (me *ContextResolver) Run(cmd *cobra.Command) (context.Context, error) {

	if me.Version {
		cmd.Printf("%s %s %s\n", version.Package, version.Version, version.Revision)
		os.Exit(0)
	}

	var level zerolog.Level
	if me.Debug {
		level = zerolog.TraceLevel
	} else if me.Quiet {
		level = zerolog.NoLevel
	} else {
		level = zerolog.InfoLevel
	}

	ctx := cmd.Context()

	ctx = szerolog.NewVerboseConsoleLogger().Level(level).WithContext(ctx)

	return ctx, nil
}
