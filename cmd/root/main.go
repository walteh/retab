package root

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/walteh/retab/cmd/root/buf"
	"github.com/walteh/retab/cmd/root/fmt"
	"github.com/walteh/retab/cmd/root/generate"
	"github.com/walteh/retab/cmd/root/hcl"
	"github.com/walteh/retab/cmd/root/install"
	"github.com/walteh/snake"

	"github.com/walteh/retab/version"
)

type Root struct {
	Quiet   bool
	Debug   bool
	Version bool
}

var _ snake.Snakeable = (*Root)(nil)

func (me *Root) BuildCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "retab",
		Short: "retab brings tabs to terraform",
	}

	cmd.PersistentFlags().BoolVarP(&me.Quiet, "quiet", "q", false, "Do not print any output")
	cmd.PersistentFlags().BoolVarP(&me.Debug, "debug", "d", false, "Print debug output")
	cmd.PersistentFlags().BoolVarP(&me.Version, "version", "v", false, "Print version and exit")

	snake.MustNewCommand(ctx, cmd, "fmt", &fmt.Handler{})
	snake.MustNewCommand(ctx, cmd, "buf", &buf.Handler{})
	snake.MustNewCommand(ctx, cmd, "hcl", &hcl.Handler{})
	snake.MustNewCommand(ctx, cmd, "install", &install.Handler{})
	snake.MustNewCommand(ctx, cmd, "generate", &generate.Handler{})

	return cmd
}

func (me *Root) ParseArguments(ctx context.Context, cmd *cobra.Command, _ []string) error {

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
