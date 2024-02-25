package format

import (
	"context"
	"io"
	"reflect"

	"github.com/rs/zerolog"
	"github.com/walteh/terrors"
)

type Provider interface {
	Format(ctx context.Context, cfg Configuration, reader io.Reader) (io.Reader, error)
	Targets() []string
}

func Format(ctx context.Context, provider Provider, cfg ConfigurationProvider, filename string, fle io.Reader) (io.Reader, error) {

	ctx = zerolog.Ctx(ctx).With().Str("path", filename).Str("provider", reflect.TypeOf(provider).Elem().String()).Logger().WithContext(ctx)

	efg, err := cfg.GetConfigurationForFileType(ctx, filename)
	if err != nil {
		return nil, terrors.Wrap(err, "failed to get editorconfig")
	}

	r, err := provider.Format(ctx, efg, fle)
	if err != nil {
		return nil, terrors.Wrap(err, "failed to format")
	}

	return r, nil
}
