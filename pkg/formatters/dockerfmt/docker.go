package dockerfmt

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/walteh/retab/v2/pkg/format"
	"gitlab.com/tozd/go/errors"
)

// Formatter implements the format.Provider interface for Dockerfiles.
type Formatter struct {
}

var _ format.Provider = (*Formatter)(nil)

// NewFormatter creates a new Dockerfile formatter.
func NewFormatter() *Formatter {
	return &Formatter{}
}

// Format parses and formats Dockerfile code.
func (f *Formatter) Format(ctx context.Context, cfg format.Configuration, read io.Reader) (io.Reader, error) {

	idnt := "\t"
	if !cfg.UseTabs() && cfg.IndentSize() > 0 {
		idnt = strings.Repeat(" ", cfg.IndentSize())
	}

	dockerConfig := NewConfig(ctx, cfg)

	dockerConfig.Indent = idnt
	dockerConfig.SpaceRedirects = true
	dockerConfig.TrailingNewline = true

	// Format the Dockerfile
	formattedContent, err := FormatFileLines(read, dockerConfig)
	if err != nil {
		return nil, errors.Errorf("failed to format Dockerfile: %w", err)
	}

	return bytes.NewReader([]byte(formattedContent)), nil
}

// getBoolOption is a helper function to check if a config option is true, with a default
func getBoolOption(options map[string]string, name string, defaultValue bool) bool {
	val, ok := options[name]
	if !ok {
		return defaultValue
	}
	return val == "true"
}
