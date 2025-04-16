//go:build js && wasm

package fmt

import (
	"context"
	"io"
	"strings"
	"syscall/js"

	"github.com/rs/zerolog"
	"github.com/walteh/retab/v2/pkg/editorconfig"
	"github.com/walteh/retab/v2/pkg/format"
	"github.com/walteh/retab/v2/pkg/formatters"
	"gitlab.com/tozd/go/errors"
)

var (
	lastResult string
	lastError  error
)

var cfg *formatters.AutoFormatProvider

func init() {
	cfg = NewAutoFormatConfig()
}

func Fmt(ctx context.Context, this js.Value, args []js.Value) (string, error) {
	if len(args) != 4 {
		return "", errors.New("expected 4 arguments: formatter, filename, content, editorconfig-content")
	}

	ctx, exit := trackStats(ctx)
	defer func() { exit(ctx) }()

	formatter := args[0].String()
	filename := args[1].String()
	content := args[2].String()
	editorconfigContent := args[3].String()

	ctx = applyValueToContext(ctx, "filename", filename)

	var cfgProvider format.ConfigurationProvider
	var err error
	// Setup editorconfig with either raw content or auto-resolution
	cfgProvider, err = editorconfig.NewRawConfigurationProvider(ctx, editorconfigContent)
	if err != nil {
		zerolog.Ctx(ctx).Warn().Err(err).Msg("failed to parse editorconfig content, using default configuration")
		cfgProvider = format.NewDefaultConfigurationProvider()
	}

	br := strings.NewReader(content)

	// Get the appropriate formatter
	fmtr, err := cfg.GetFormatter(ctx, formatter, filename, br)
	if err != nil {
		return "", errors.Errorf("getting formatter: %w", err)
	}

	// Format the content
	r, err := format.Format(ctx, fmtr, cfgProvider, filename, br)
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
