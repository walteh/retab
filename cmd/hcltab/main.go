package main

import (
	"context"
	"log"
	"os"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/walteh/tftab/cmd/hcltab/handler"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
}

type CLI struct {
	handler.Handler `kong:"cmd,help='Run a handler',aliases='h'"`
	Quiet           bool `kong:"short='q',help='Do not print any output',env='QUIET'"`
	Debug           bool `kong:"short='d',help='Print debug output',env='DEBUG'"`
	Version         bool `kong:"short='v',help='Print version and exit'"`
}

func run() error {

	ctx := context.Background()

	cli := CLI{}

	k := kong.Parse(&cli, kong.Name("hcltab"))

	if cli.Version {
		kong.DefaultHelpPrinter(kong.HelpOptions{}, k)
		return nil
	}

	if cli.Quiet {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	} else if cli.Debug {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	ctx = zerolog.New(os.Stderr).With().Timestamp().Logger().With().Str("app", "hcltab").Logger().WithContext(ctx)

	k.BindTo(ctx, (*context.Context)(nil))
	k.BindTo(afero.NewOsFs(), (*afero.Fs)(nil))

	err := k.Run(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("pipeline failed")
		return err
	}

	return nil

}

func main() {
	if run() != nil {
		os.Exit(1)
	}
}
