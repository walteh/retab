package externalwrite

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/format"
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

type externalFormatter struct {
	internal ExternalFormatter
}

func (me *externalFormatter) Targets() []string {
	return me.internal.Targets()
}

func ExternalFormatterToProvider(ext ExternalFormatter) format.Provider {
	return &externalFormatter{ext}
}

func (me *externalFormatter) Format(ctx context.Context, cfg configuration.Provider, input io.Reader) (io.Reader, error) {

	read, f := me.internal.Format(ctx, input)

	var rerr error

	go func() {
		if err := f(); err != nil {
			rerr = err
		}
	}()

	output, err := applyConfiguration(ctx, me.internal, cfg, read)
	if err != nil {
		return nil, err
	}

	if rerr != nil {
		return nil, rerr
	}

	return output, nil
}

func applyConfiguration(ctx context.Context, ext ExternalFormatter, cfg configuration.Provider, input io.Reader) (io.Reader, error) {
	var output bytes.Buffer
	scanner := bufio.NewScanner(input)
	indentation := "\t"
	if !cfg.UseTabs() {
		indentation = strings.Repeat(" ", cfg.IndentSize())
	}

	previousLineWasEmpty := false
	for scanner.Scan() {
		line := scanner.Text()

		// Apply indentation preference.
		line = strings.Replace(line, ext.Indent(), indentation, -1)

		// Trim multiple empty lines if configured.
		if cfg.TrimMultipleEmptyLines() {
			if strings.TrimSpace(line) == "" {
				if previousLineWasEmpty {
					continue
				}
				previousLineWasEmpty = true
			} else {
				previousLineWasEmpty = false
			}
		}

		// Write the modified line to the output buffer.
		_, err := output.WriteString(line + "\n")
		if err != nil {
			return nil, err
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &output, nil
}
