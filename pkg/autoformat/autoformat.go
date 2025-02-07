package autoformat

import (
	"context"
	"io"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/walteh/retab/v2/pkg/format"
	"github.com/walteh/retab/v2/pkg/format/cmdfmt"
	"github.com/walteh/retab/v2/pkg/format/hclfmt"
	"github.com/walteh/retab/v2/pkg/format/protofmt"
	"gitlab.com/tozd/go/errors"
)

// GetFormatter returns the appropriate formatter for the given format type and filename
func GetFormatter(formatType string) (format.Provider, error) {
	var fmtr format.Provider

	if formatType == "auto" {
		return nil, nil // Caller should handle auto-detection
	} else if formatType == "hcl" {
		fmtr = hclfmt.NewFormatter()
	} else if formatType == "proto" {
		fmtr = protofmt.NewFormatter()
	} else if formatType == "dart" {
		fmtr = cmdfmt.NewDartFormatter("dart")
	} else if formatType == "tf" {
		fmtr = cmdfmt.NewTerraformFormatter("terraform")
	} else {
		return nil, errors.Errorf("invalid formatter type: %q", formatType)
	}

	return fmtr, nil
}

// AutoDetectFormatter attempts to find a suitable formatter based on the filename
func AutoDetectFormatter(filename string) (format.Provider, error) {
	pfmters := []format.Provider{
		hclfmt.NewFormatter(),
		protofmt.NewFormatter(),
		cmdfmt.NewDartFormatter("dart"),
		cmdfmt.NewTerraformFormatter("terraform"),
	}

	basename := filepath.Base(filename)
	for _, pfmtr := range pfmters {
		for _, target := range pfmtr.Targets() {
			ok, err := doublestar.Match(target, basename)
			if err != nil {
				return nil, errors.Errorf("failed to match glob: %w", err)
			}
			if ok {
				return pfmtr, nil
			}
		}
	}

	return nil, nil
}

// FormatFile handles the common formatting logic for both CLI and WASM
func FormatFile(ctx context.Context, formatType string, filename string, input io.Reader, cfg format.ConfigurationProvider) (io.Reader, error) {
	var fmtr format.Provider
	var err error

	if formatType == "auto" {
		fmtr, err = AutoDetectFormatter(filename)
		if err != nil {
			return nil, errors.Errorf("auto-detecting formatter: %w", err)
		}
	} else {
		fmtr, err = GetFormatter(formatType)
		if err != nil {
			return nil, err
		}
	}

	if fmtr == nil {
		return nil, errors.Errorf("no formatters found for file %q", filename)
	}

	r, err := format.Format(ctx, fmtr, cfg, filename, input)
	if err != nil {
		return nil, errors.Errorf("formatting file: %w", err)
	}

	return r, nil
}
