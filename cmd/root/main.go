package root

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/walteh/tftab/cmd/root/tftab"
	"github.com/walteh/tftab/pkg/cli"
)

type Root struct {
	Quiet   bool
	Debug   bool
	Version bool
}

func (me *Root) Define(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tftab",
		Short: "tftab brings tabs to terraform",
	}

	cmd.PersistentFlags().BoolVarP(&me.Quiet, "quiet", "q", false, "Do not print any output")
	cmd.PersistentFlags().BoolVarP(&me.Debug, "debug", "d", false, "Print debug output")
	cmd.PersistentFlags().BoolVarP(&me.Version, "version", "v", false, "Print version and exit")

	cli.RegisterCommand(ctx, cmd, &tftab.Handler{})

	return cmd
}

func (me *Root) InjectContext(cmd *cobra.Command) (context.Context, error) {

	var level zerolog.Level
	if me.Debug {
		level = zerolog.TraceLevel
	} else if me.Quiet {
		level = zerolog.NoLevel
	} else {
		level = zerolog.InfoLevel
	}

	ctx := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger().Level(level).WithContext(cmd.Context())

	if me.Version {
		cmd.Println("tftab version 0.0.1")
		os.Exit(0)
	}

	return ctx, nil
}
