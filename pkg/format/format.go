package format

import (
	"context"
	"io"
	"path/filepath"
	"reflect"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rs/zerolog"
	"gitlab.com/tozd/go/errors"
)

type Provider interface {
	Format(ctx context.Context, cfg Configuration, reader io.Reader) (io.Reader, error)
	Targets() []string
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

// AutoDetectFormatter attempts to find a suitable formatter based on the filename
func AutoDetectFormatter(filename string, formatters []Provider) (Provider, error) {
	basename := filepath.Base(filename)
	for _, fmtr := range formatters {
		for _, target := range fmtr.Targets() {
			ok, err := doublestar.Match(target, basename)
			if err != nil {
				return nil, errors.Errorf("failed to match glob: %w", err)
			}
			if ok {
				return fmtr, nil
			}
		}
	}

	return nil, nil
}
