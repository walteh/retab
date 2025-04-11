package cmdfmt

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/walteh/retab/v2/pkg/format"
	"gitlab.com/tozd/go/errors"
)

type ExternalFormatterConfig struct {
	Indentation string
	Targets     []string
}

type ExternalFormatter interface {
	Format(ctx context.Context, reader io.Reader) (io.Reader, func() error)
	Indent() string
	Targets() []string
}

type externalStdinFormatter struct {
	internal ExternalFormatter
}

func (me *externalStdinFormatter) Targets() []string {
	return me.internal.Targets()
}

func ExternalFormatterToProvider(ext ExternalFormatter) format.Provider {
	return &externalStdinFormatter{ext}
}

func (me *externalStdinFormatter) Format(ctx context.Context, cfg format.Configuration, input io.Reader) (io.Reader, error) {

	read, f := me.internal.Format(ctx, input)

	var rerr error
	go func() {
		if err := f(); err != nil {
			rerr = err
		}
	}()

	output, err := applyConfiguration(ctx, me.internal, cfg, read)
	if err != nil {
		return nil, errors.Errorf("failed to apply configuration: %w", err)
	}

	if rerr != nil {
		return nil, errors.Errorf("failed to format: %w", rerr)
	}

	return output, nil
}

// type externalFileFormatter struct {
// 	internal ExternalFileFormatter
// 	fmter    ExternalFormatter
// }

// func (me *externalFileFormatter) Targets() []string {
// 	return me.internal.Targets()
// }

// func ExternalFileFormatterToProvider(ext ExternalFormatter) format.Provider {
// 	return &externalStdinFormatter{ext}
// }

func applyConfiguration(_ context.Context, ext ExternalFormatter, cfg format.Configuration, input io.Reader) (io.Reader, error) {
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

		// if !someOutput && line != "" {
		// 	someOutput = true
		// }

		// Apply indentation preference.
		line = strings.Replace(line, ext.Indent(), indentation, -1)

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
