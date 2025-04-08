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

	// Do the replacements after formatting
	result := buf.String()
	for id, value := range fmtr.replacers {
		result = strings.Replace(result, id, value, -1)
	}

	return strings.NewReader(result), nil
}

// func (me *Formatter) FormatExperimental(ctx context.Context, cfg format.Configuration, read io.Reader) (io.Reader, error) {
// 	readBytes, err := io.ReadAll(read)
// 	if err != nil {
// 		return nil, errors.Errorf("failed to read protobuf: %w", err)
// 	}

// 	reportf := report.NewFile("retab.protobuf-parser", string(readBytes))

// 	reports := &report.Report{}
// 	fileNode, ok := parser.Parse(reportf, reports)
// 	if !ok {
// 		reports.Canonicalize()
// 		ren := report.Renderer{
// 			Compact: true,
// 		}
// 		strw := &strings.Builder{}
// 		errs, warns, err := ren.Render(reports, strw)
// 		if err != nil {
// 			return nil, errors.Errorf("failed to render protobuf report: %w", err)
// 		}

// 		return nil, errors.Errorf("failed to parse protobuf errors[len=%d] warnings[len=%d]: %w", errs, warns, strw.String())
// 	}

// 	var buf bytes.Buffer
// 	fmtr := newFormatter(&buf, fileNode, cfg)

// 	if err := fmtr.Run(); err != nil {
// 		return nil, errors.Errorf("failed to format: %w", err)
// 	}

// 	// Do the replacements after formatting
// 	result := buf.String()
// 	for id, value := range fmtr.replacers {
// 		result = strings.Replace(result, id, value, -1)
// 	}

// 	return strings.NewReader(result), nil
// }
