package format

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"text/tabwriter"

	"github.com/rs/zerolog"
	"gitlab.com/tozd/go/errors"
)

type Provider interface {
	Format(ctx context.Context, cfg Configuration, reader io.Reader) (io.Reader, error)
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

func FormatSimple(ctx context.Context, provider Provider, filename string, useTabs bool, indentSize int, input io.Reader) (io.Reader, error) {
	return Format(ctx, provider, &basicConfigurationProvider{
		tabs:       useTabs,
		indentSize: indentSize,
		filename:   filename,
	}, filename, input)
}

func FormatSimpleBytes(ctx context.Context, provider Provider, filename string, useTabs bool, indentSize int, input []byte) ([]byte, error) {
	r, err := FormatSimple(ctx, provider, filename, useTabs, indentSize, bytes.NewReader(input))
	if err != nil {
		return nil, errors.Errorf("failed to format: %w", err)
	}
	return io.ReadAll(r)
}

func BuildTabWriter(writer io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(writer, 0, 1, 1, ' ', tabwriter.TabIndent|tabwriter.StripEscape|tabwriter.DiscardEmptyColumns)
}
