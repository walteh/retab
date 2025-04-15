//go:build !js

package fmt

// `hclFmt` command recursively looks for hcl files in the directory tree starting at workingDir, and formats them
// based on the language style guides provided by Hashicorp. This is done using the official hcl2 library.

import (
	"context"
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/walteh/retab/v2/pkg/format"
	"github.com/walteh/retab/v2/pkg/format/editorconfig"
	"gitlab.com/tozd/go/errors"
)

type Handler struct {
	filename            string
	formatter           string // auto, hcl, proto, dart, tf
	ToStdout            bool
	FromStdin           bool
	editorconfigContent string
}

func NewFmtCommand() *cobra.Command {
	me := &Handler{}

	cmd := &cobra.Command{
		Use:   "fmt",
		Short: "format files with the hcl golang library, but with tabs",
	}

	cmd.Flags().StringVar(&me.formatter, "formatter", "auto", "the formatter to use")
	cmd.Flags().BoolVar(&me.ToStdout, "stdout", false, "write to stdout instead of file")
	cmd.Flags().BoolVar(&me.FromStdin, "stdin", false, "read from stdin instead of file")
	cmd.Flags().StringVar(&me.editorconfigContent, "editorconfig-content", "", "editorconfig content (optional)")
	cmd.Args = cobra.ExactArgs(1)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		me.filename = args[0]
		return me.Run(cmd.Context())
	}

	return cmd
}

func (me *Handler) Run(ctx context.Context) error {

	defer trackStats(ctx)()

	var err error
	var cfgProvider format.ConfigurationProvider
	// Setup editorconfig with either raw content or auto-resolution
	cfgProvider, err = editorconfig.NewRawConfigurationProvider(ctx, me.editorconfigContent)
	if err != nil {
		zerolog.Ctx(ctx).Warn().Err(err).Msg("failed to parse editorconfig content, using default configuration")
		cfgProvider = format.NewDefaultConfigurationProvider()
	}

	var input io.ReadSeeker
	if me.FromStdin {
		input = os.Stdin
	} else {
		file, err := os.Open(me.filename)
		if err != nil {
			return errors.Errorf("opening file: %w", err)
		}
		defer file.Close()
		input = file
	}

	fmtr, err := getFormatter(ctx, me.formatter, me.filename, input)
	if err != nil {
		return err
	}

	r, err := format.Format(ctx, fmtr, cfgProvider, me.filename, input)
	if err != nil {
		return errors.Errorf("formatting content: %w", err)
	}

	if me.ToStdout || me.FromStdin {
		_, err = io.Copy(os.Stdout, r)
		return err
	}

	rBytes, err := io.ReadAll(r)
	if err != nil {
		return errors.Errorf("reading formatted content: %w", err)
	}

	err = os.WriteFile(me.filename, rBytes, 0644)
	if err != nil {
		return errors.Errorf("writing formatted file: %w", err)
	}

	return nil
}
