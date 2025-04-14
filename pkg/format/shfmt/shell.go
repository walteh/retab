package shfmt

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"strings"

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

// func GetShebangFromBytes(bs []byte) (string, error) {
// 	return fileutil.Shebang(bs), nil
// }

// func GetShebangFromReader(read io.Reader) (string, io.Reader, error) {

// 	var buf bytes.Buffer
// 	tee := io.TeeReader(read, &buf)
// 	bs, err := io.ReadAll(tee)
// 	if err != nil {
// 		return "", &buf, errors.Errorf("failed to read shell script: %w", err)
// 	}

// 	read = &buf

// 	langs := enry.GetLanguagesByShebang("", bs, nil)
// 	if len(langs) == 0 {
// 		return "", &buf, errors.Errorf("failed to get language")
// 	}

// 	lang := langs[0]

// 	shebasng := fileutil.Shebang(bs)

// 	return lang, shebasng, read, nil
// }

// Format parses and formats shell code.
func (f *Formatter) Format(ctx context.Context, cfg format.Configuration, read io.Reader) (io.Reader, error) {

	langVar := syntax.LangAuto

	if dialect, ok := cfg.Raw()["shell_dialect"]; ok {
		if err := langVar.Set(dialect); err != nil {
			return nil, errors.Errorf("invalid shell dialect %q: %w", dialect, err)
		}
	} else {
		readz := bufio.NewReader(read)
		cont, _ := readz.Peek(250)
		read = readz

		lang := format.Shebang(cont)

		switch lang {
		case "bash", "zsh", "ksh":
			// Bash-compatible shells with extended features
			langVar = syntax.LangBash
			break
		case "posix", "sh", "dash", "yash", "ash", "busybox":
			// POSIX-compliant shells (standard/basic shell)
			langVar = syntax.LangPOSIX
			break
		case "mksh", "pdksh":
			// MirBSD Korn shell
			langVar = syntax.LangMirBSDKorn
			break
		case "bats":
			// Bash Automated Testing System
			langVar = syntax.LangBats
			break
		default:
			// Skip unknown shell types
			langVar = syntax.LangBash
		}

	}

	// Create a parser that keeps comments
	parser := syntax.NewParser(syntax.KeepComments(true), syntax.Variant(langVar))

	// Parse the source code
	prog, err := parser.Parse(read, cfg.Raw()["filename"])
	if err != nil {
		return nil, errors.Errorf("failed to parse shell script: %w", err)
	}

	// Create a new printer
	printer := syntax.NewPrinter()

	// Apply configuration
	var indent uint
	if !cfg.UseTabs() {
		indent = uint(cfg.IndentSize())
	} else {
		indent = 4
		// we could set this to 0, but we need to hack it with the brute force
		// indentation below
	}

	syntax.FunctionNextLine(false)(printer)
	syntax.SwitchCaseIndent(true)(printer)
	syntax.SpaceRedirects(true)(printer)
	// syntax.KeepPadding(false)(printer)
	syntax.BinaryNextLine(true)(printer)
	syntax.Indent(indent)(printer)
	syntax.Minify(false)(printer)
	syntax.SingleLine(false)(printer)
	// syntax.Simplify(prog)

	// Format the code
	var buf bytes.Buffer
	// wrt := format.BuildTabWriter(&buf)
	err = printer.Print(&buf, prog)
	if err != nil {
		return nil, errors.Errorf("failed to format shell script: %w", err)
	}

	if cfg.UseTabs() {
		//replace all leading 4 spaces with tabs
		// the way the tab writer is configured inside syntax.NewPrinter() makes
		// comment alignment way off unless we hack it like this
		br, err := format.BruteForceIndentation(ctx, strings.Repeat(" ", 4), cfg, &buf)
		if err != nil {
			return nil, errors.Errorf("failed to apply configuration: %w", err)
		}

		return br, nil
	}

	return bytes.NewReader(buf.Bytes()), nil
}
