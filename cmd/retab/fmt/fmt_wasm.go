//go:build js && wasm

package fmt

import (
	"context"
	"io"
	"strings"
	"syscall/js"

	"github.com/rs/zerolog"
	"github.com/walteh/retab/v2/pkg/format"
	"github.com/walteh/retab/v2/pkg/format/editorconfig"
	"gitlab.com/tozd/go/errors"
)

var (
	lastResult string
	lastError  error
)

func Fmt(ctx context.Context, this js.Value, args []js.Value) (string, error) {
	if len(args) != 4 {
		return "", errors.New("expected 4 arguments: formatter, filename, content, editorconfig-content")
	}

	zerolog.Ctx(ctx).Info().Msg("fmt")

	formatter := args[0].String()
	filename := args[1].String()
	content := args[2].String()
	editorconfigContent := args[3].String()
	var cfgProvider format.ConfigurationProvider
	var err error
	// Setup editorconfig with either raw content or auto-resolution
	cfgProvider, err = editorconfig.NewRawConfigurationProvider(ctx, editorconfigContent)
	if err != nil {
		zerolog.Ctx(ctx).Warn().Err(err).Msg("failed to parse editorconfig content, using default configuration")
		cfgProvider = format.NewDefaultConfigurationProvider()
	}

	re := strings.NewReader(content)

	// Get the appropriate formatter
	fmtr, err := getFormatter(ctx, formatter, filename)
	if err != nil {
		return "", errors.Errorf("getting formatter: %w", err)
	}

	// Format the content
	r, err := format.Format(ctx, fmtr, cfgProvider, filename, re)
	if err != nil {
		return "", errors.Errorf("formatting content: %w", err)
	}

	// Read the formatted content
	result, err := io.ReadAll(r)
	if err != nil {
		return "", errors.Errorf("reading formatted content: %w", err)
	}

	return string(result), nil
}
