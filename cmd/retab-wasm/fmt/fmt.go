//go:build js && wasm

package fmt

import (
	"context"
	"io"
	"strings"
	"syscall/js"

	"github.com/walteh/retab/v2/pkg/format"
	"github.com/walteh/retab/v2/pkg/format/cmdfmt"
	"github.com/walteh/retab/v2/pkg/format/editorconfig"
	"github.com/walteh/retab/v2/pkg/format/hclfmt"
	"github.com/walteh/retab/v2/pkg/format/protofmt"
	"gitlab.com/tozd/go/errors"
)

var (
	lastResult string
	lastError  error
)

func getFormatter(formatType string, filename string) (format.Provider, error) {
	if formatType == "auto" {
		formatters := []format.Provider{
			hclfmt.NewFormatter(),
			protofmt.NewFormatter(),
			cmdfmt.NewDartFormatter("dart"),
			cmdfmt.NewTerraformFormatter("terraform"),
		}
		fmtr, err := format.AutoDetectFormatter(filename, formatters)
		if err != nil {
			return nil, errors.Errorf("auto-detecting formatter: %w", err)
		}
		if fmtr == nil {
			return nil, errors.Errorf("no formatters found for file %q", filename)
		}
		return fmtr, nil
	}

	switch formatType {
	case "hcl":
		return hclfmt.NewFormatter(), nil
	case "proto":
		return protofmt.NewFormatter(), nil
	case "dart":
		return cmdfmt.NewDartFormatter("dart"), nil
	case "tf":
		return cmdfmt.NewTerraformFormatter("terraform"), nil
	default:
		return nil, errors.Errorf("invalid formatter type: %q", formatType)
	}
}

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
	fmtr, err := getFormatter(formatter, filename)
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
