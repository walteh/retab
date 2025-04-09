package shfmt

import (
	"bytes"
	"context"
	"io"

	"github.com/walteh/retab/v2/pkg/format"

	"gitlab.com/tozd/go/errors"
	"mvdan.cc/sh/v3/syntax"
)

// Formatter implements the format.Provider interface for shell scripts.
type Formatter struct {
}

var _ format.Provider = (*Formatter)(nil)

// NewFormatter creates a new shell formatter.
func NewFormatter() *Formatter {
	return &Formatter{}
}

// Targets returns the file patterns this formatter handles.
func (f *Formatter) Targets() []string {
	return []string{"*.sh", "*.bash", "*.ksh", "*.zsh", "*.bats"}
}

// Format parses and formats shell code.
func (f *Formatter) Format(ctx context.Context, cfg format.Configuration, read io.Reader) (io.Reader, error) {
	// Determine the shell dialect based on configuration or file extension
	// Default to bash if not specified
	langVar := syntax.LangBash

	if dialect, ok := cfg.Raw()["shell_dialect"]; ok {
		if err := langVar.Set(dialect); err != nil {
			return nil, errors.Errorf("invalid shell dialect %q: %w", dialect, err)
		}
	}

	// Create a parser that keeps comments
	parser := syntax.NewParser(syntax.KeepComments(true), syntax.Variant(langVar))

	// Parse the source code
	prog, err := parser.Parse(read, "")
	if err != nil {
		return nil, errors.Errorf("failed to parse shell script: %w", err)
	}

	// Create a new printer
	printer := syntax.NewPrinter()

	// // Apply minify setting if present
	// minify := cfg.Raw()["minify"] == "true"
	// if minify {
	// 	syntax.Minify(true)(printer)
	// }

	// Apply configuration
	var indent uint
	if !cfg.UseTabs() {
		indent = uint(cfg.IndentSize())
	}

	syntax.FunctionNextLine(false)(printer)
	syntax.SwitchCaseIndent(true)(printer)
	syntax.SpaceRedirects(true)(printer)
	syntax.KeepPadding(true)(printer)
	syntax.BinaryNextLine(true)(printer)
	syntax.Indent(indent)(printer)
	syntax.Minify(false)(printer)
	syntax.SingleLine(false)(printer)
	// syntax.Simplify(prog)
	syntax.SpaceRedirects(true)(printer)

	// Format the code
	var buf bytes.Buffer
	err = printer.Print(&buf, prog)
	if err != nil {
		return nil, errors.Errorf("failed to format shell script: %w", err)
	}

	return bytes.NewReader(buf.Bytes()), nil
}
