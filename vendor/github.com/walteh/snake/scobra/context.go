package scobra

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/walteh/snake/szerolog"
)

type ContextResolver struct {
	Quiet bool `usage:"Do not print any output" default:"false"`
	Debug bool `usage:"Print debug output" default:"false"`
}

func (me *ContextResolver) Run(cmd *cobra.Command) (context.Context, error) {

	var level zerolog.Level
	if me.Debug {
		level = zerolog.TraceLevel
	} else if me.Quiet {
		level = zerolog.NoLevel
	} else {
		level = zerolog.InfoLevel
	}

	ctx := context.Background()

	ctx = szerolog.NewConsoleLoggerContext(ctx, level, cmd.OutOrStdout())

	return ctx, nil
}
