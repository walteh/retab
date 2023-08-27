package root

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/walteh/snake"
	"github.com/walteh/tftab/cmd/root/buf"
	"github.com/walteh/tftab/cmd/root/fmt"
	"github.com/walteh/tftab/version"
)

type Root struct {
	Quiet   bool
	Debug   bool
	Version bool
}

var _ snake.Snakeable = (*Root)(nil)

func (me *Root) BuildCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tftab",
		Short: "tftab brings tabs to terraform",
	}

	cmd.PersistentFlags().BoolVarP(&me.Quiet, "quiet", "q", false, "Do not print any output")
	cmd.PersistentFlags().BoolVarP(&me.Debug, "debug", "d", false, "Print debug output")
	cmd.PersistentFlags().BoolVarP(&me.Version, "version", "v", false, "Print version and exit")

	snake.MustNewCommand(ctx, cmd, "fmt", &fmt.Handler{})
	snake.MustNewCommand(ctx, cmd, "buf", &buf.Handler{})

	return cmd
}

func (me *Root) ParseArguments(ctx context.Context, cmd *cobra.Command, args []string) error {

	var level zerolog.Level
	if me.Debug {
		level = zerolog.TraceLevel
	} else if me.Quiet {
		level = zerolog.NoLevel
	} else {
		level = zerolog.InfoLevel
	}

	ctx = zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger().Level(level).WithContext(ctx)

	if me.Version {
		cmd.Printf("%s %s %s\n", version.Package, version.Version, version.Revision)
		os.Exit(0)
	}

	cmd.SetContext(ctx)

	return nil
}
