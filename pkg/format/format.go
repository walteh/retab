package format

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/rs/zerolog"
	"gitlab.com/tozd/go/errors"
)

type Provider interface {
	Format(ctx context.Context, cfg Configuration, reader io.Reader) (io.Reader, error)
	// Targets() []string
}

func Format(ctx context.Context, provider Provider, cfg ConfigurationProvider, filename string, fle io.Reader) (io.Reader, error) {
	ctx = zerolog.Ctx(ctx).With().Str("path", filename).Str("provider", reflect.TypeOf(provider).Elem().String()).Logger().WithContext(ctx)

	efg, err := cfg.GetConfigurationForFileType(ctx, filename)
	if err != nil {
		return nil, errors.Errorf("failed to get editorconfig: %w", err)
	}

	r, err := provider.Format(ctx, efg, fle)
	if err != nil {
		return nil, errors.Errorf("failed to format: %w", err)
	}

	return r, nil
}

func BuildTabWriter(writer io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(writer, 0, 1, 1, ' ', tabwriter.TabIndent|tabwriter.StripEscape|tabwriter.DiscardEmptyColumns)
}

func BruteForceIndentation(ctx context.Context, startIndentation string, cfg Configuration, input io.Reader) (io.Reader, error) {
	var output bytes.Buffer
	scanner := bufio.NewScanner(input)
	indentation := "\t"
	if !cfg.UseTabs() {
		indentation = strings.Repeat(" ", cfg.IndentSize())
	}

	previousLineWasEmpty := false
	// someOutput := false
	for scanner.Scan() {
		line := scanner.Text()

		newLine := ""

		// Apply indentation preference.
		for strings.HasPrefix(line, startIndentation) {
			newLine += indentation
			line = line[len(startIndentation):]
		}

		line = newLine + line

		// ==========================================
		// this trims multiple empty lines, but leaves single empty lines
		// wrapping like this just to make it explicity clear what code
		// 		is responsible for this in case we need to disable it
		if strings.TrimSpace(line) == "" {
			if previousLineWasEmpty {
				continue
			}
			previousLineWasEmpty = true
		} else {
			previousLineWasEmpty = false
		}
		// ==========================================

		// Write the modified line to the output buffer.
		_, err := output.WriteString(line + "\n")
		if err != nil {
			return nil, errors.Errorf("failed to write to output buffer: %w", err)
		}
	}

	// if !someOutput {
	// 	return nil,terrors.Errorf("no output from external formatter")
	// }

	if err := scanner.Err(); err != nil {
		failString := "failed to read output from external formatter"
		outputStr := output.String()
		if outputStr != "" {
			failString = failString + ": " + outputStr
		}
		return nil, errors.Errorf("%s: %w", failString, err)
	}

	return &output, nil
}
