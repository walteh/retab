package dockerfmt

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/reteps/dockerfmt/lib"
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

// Targets returns the file patterns this formatter handles.
func (f *Formatter) Targets() []string {
	return []string{"Dockerfile", "Dockerfile.*", "*.dockerfile"}
}

// Format parses and formats Dockerfile code.
func (f *Formatter) Format(ctx context.Context, cfg format.Configuration, read io.Reader) (io.Reader, error) {
	// Read all content from the reader
	content, err := io.ReadAll(read)
	if err != nil {
		return nil, errors.Errorf("failed to read Dockerfile content: %w", err)
	}

	// Split the content string by newlines, ensuring trailing newlines are preserved
	contentStr := string(content)
	// endsWithNewline := strings.HasSuffix(contentStr, "\n")

	// Split content into lines, maintaining line endings
	lines := strings.SplitAfter(contentStr, "\n")

	indentSize := cfg.IndentSize()
	if cfg.UseTabs() {
		indentSize *= 4
	}

	// Create dockerfmt configuration
	dockerConfig := &lib.Config{
		IndentSize: uint(indentSize),
		// TrailingNewline: getBoolOption(cfg.Raw(), "trailing_newline", true),
		// SpaceRedirects:  getBoolOption(cfg.Raw(), "space_redirects", false),
		SpaceRedirects:  true,
		TrailingNewline: true,
	}

	// Format the Dockerfile
	formattedContent := lib.FormatFileLines(lines, dockerConfig)

	// // Ensure we maintain the original trailing newline state
	// if endsWithNewline && !strings.HasSuffix(formattedContent, "\n") {
	// 	formattedContent += "\n"
	// } else if !endsWithNewline && strings.HasSuffix(formattedContent, "\n") {
	// 	formattedContent = strings.TrimSuffix(formattedContent, "\n")
	// }

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
