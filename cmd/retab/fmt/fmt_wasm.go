//go:build js && wasm

package fmt

import (
	"context"
	"io"
	"strings"
	"syscall/js"

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

	formatter := args[0].String()
	filename := args[1].String()
	content := args[2].String()
	editorconfigContent := args[3].String()

	// Setup editorconfig with either raw content or auto-resolution
	cfgProvider, err := editorconfig.NewDynamicConfigurationProvider(ctx, editorconfigContent)
	if err != nil {
		return "", errors.Errorf("creating configuration provider: %w", err)
	}

	// Get the appropriate formatter
	fmtr, err := getFormatter(ctx, formatter, filename)
	if err != nil {
		return "", errors.Errorf("getting formatter: %w", err)
	}

	// Format the content
	r, err := format.Format(ctx, fmtr, cfgProvider, filename, strings.NewReader(content))
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
