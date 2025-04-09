package protofmt

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/walteh/retab/v2/pkg/format"

	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
	"gitlab.com/tozd/go/errors"
)

type Formatter struct {
}

var _ format.Provider = (*Formatter)(nil)

func NewFormatter() *Formatter {
	return &Formatter{}
}

func (me *Formatter) Targets() []string {
	return []string{"*.proto", "*.proto3"}
}

func (me *Formatter) Format(ctx context.Context, cfg format.Configuration, read io.Reader) (io.Reader, error) {
	fileNode, err := parser.Parse("retab.protobuf-parser", read, reporter.NewHandler(nil))
	if err != nil {
		return nil, errors.Errorf("failed to parse protobuf: %w", err)
	}

	var buf bytes.Buffer
	fmtr := newFormatter(&buf, fileNode, cfg)

	if err := fmtr.Run(); err != nil {
		return nil, errors.Errorf("failed to format: %w", err)
	}

	result := buf.String()

	for _, replacement := range fmtr.replacers {
		result = strings.Replace(result, replacement.id, replacement.new, -1)
		// TODO(fix): we could remove the trailing whitespace here if we want to
		// we can't do it in the formatter because it needs to be done after the replacements are injected
	}
	if cfg.UseTabs() {
		result = strings.ReplaceAll(result, "$indent$", "\t")
	} else {
		result = strings.ReplaceAll(result, "$indent$", strings.Repeat(" ", cfg.IndentSize()))
	}

	return strings.NewReader(result), nil
}

func (f *formatter) inspectTabWriter() string {
	// Create a new buffer to capture output
	var buf bytes.Buffer

	// Create a new tabwriter with the same settings
	inspector := format.BuildTabWriter(f.cfg, &buf)

	// Get the current content by copying from the original writer's buffer
	// This assumes format.BuildTabWriter uses the same settings
	if original, ok := f.writer.(*bytes.Buffer); ok {
		// Copy the current content to the inspector
		inspector.Write(original.Bytes())
	}

	// Flush the inspector (not the original)
	inspector.Flush()

	return buf.String()
}
