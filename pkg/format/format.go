package format

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/terrors"
)

type Provider interface {
	Format(ctx context.Context, cfg configuration.Configuration, reader io.Reader) (io.Reader, error)
	Targets() []string
}

func Format(ctx context.Context, provider Provider, cfg configuration.Provider, fls afero.Fs, fle afero.File) error {

	ctx = zerolog.Ctx(ctx).With().Str("path", fle.Name()).Str("provider", reflect.TypeOf(provider).Elem().String()).Logger().WithContext(ctx)

	isdir, err := afero.IsDir(fls, fle.Name())
	if err != nil {
		return terrors.Wrap(err, "failed to check if path is a directory")
	}

	files := []string{}

	if isdir {
		zerolog.Ctx(ctx).Debug().Msg("Path is a directory. Globbing files.")
		glbfls := afero.NewBasePathFs(fls, fle.Name())
		for _, ext := range provider.Targets() {
			glb, err := afero.Glob(glbfls, ext)
			if err != nil {
				return terrors.Wrap(err, "failed to glob").Event(func(e *zerolog.Event) *zerolog.Event {
					return e.Str("pattern", ext)
				})
			}
			files = append(files, glb...)
		}
	} else {
		files = append(files, fle.Name())
	}

	if len(files) == 0 {
		zerolog.Ctx(ctx).Debug().Msg("No files to format.")
		return nil
	}

	zerolog.Ctx(ctx).Debug().Any("files", files).Msg("Formatting files.")

	grp := sync.WaitGroup{}

	var formatErrors *multierror.Error
	for _, filename := range files {
		grp.Add(1)
		go func(filename string) {
			defer grp.Done()

			datafunc := func(e *zerolog.Event) *zerolog.Event {
				return e.Str("path", filename)
			}

			fle, err := afero.ReadFile(fls, filename)
			if err != nil {
				formatErrors = multierror.Append(formatErrors, terrors.Wrap(err, "failed to open file").Event(datafunc))
				return
			}

			efg, err := cfg.GetConfigurationForFileType(ctx, filename)
			if err != nil {
				formatErrors = multierror.Append(formatErrors, terrors.Wrap(err, "failed to get editorconfig").Event(datafunc))
				return
			}

			r, err := provider.Format(ctx, efg, bytes.NewReader(fle))
			if err != nil {
				formatErrors = multierror.Append(formatErrors, terrors.Wrap(err, "failed to format file").Event(datafunc))
				return
			}

			err = afero.WriteReader(fls, filename, r)
			if err != nil {
				formatErrors = multierror.Append(formatErrors, terrors.Wrap(err, "failed to write formatted file").Event(datafunc))
				return
			}

			zerolog.Ctx(ctx).Info().Str("path", filename).Msg("formatted")
		}(filename)
	}

	grp.Wait()

	return formatErrors.ErrorOrNil()
}
