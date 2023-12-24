package format

import (
	"context"
	"io"
	"reflect"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/walteh/retab/pkg/configuration"
)

type Provider interface {
	Format(ctx context.Context, cfg configuration.Provider, reader io.Reader) (io.Reader, error)
	Targets() []string
}

func Format(ctx context.Context, provider Provider, cfg configuration.Provider, fls afero.Fs, fle afero.File) error {

	ctx = zerolog.Ctx(ctx).With().Str("provider", reflect.TypeOf(provider).String()).Logger().WithContext(ctx)

	isdir, err := afero.IsDir(fls, fle.Name())
	if err != nil {
		return err
	}

	files := []string{}

	if isdir {
		fls = afero.NewBasePathFs(fls, fle.Name())
		for _, ext := range provider.Targets() {
			glb, err := afero.Glob(fls, ext)
			if err != nil {
				return err
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

	// // handle when option specifies a particular file
	// if !isDir {

	// 	if !filepath.IsAbs(path) {
	// 		path = filepath.Join(working, path)
	// 	}

	// 	zerolog.Ctx(ctx).Debug().Msgf("Formatting hcl file at: %s.", path)

	// 	fle, err := fs.Open(path)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	r, err := provider.Format(ctx, cfg, fle)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	err = afero.WriteReader(fs, path, r)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	zerolog.Ctx(ctx).Debug().Msgf("Formatted file at: %s.", path)

	// 	return nil

	// }

	zerolog.Ctx(ctx).Debug().Any("files", files).Msg("Formatting files.")

	// afero.Glob(fls, path)

	// // zglob normalizes paths to "/"
	// var files []string

	// for _, ext := range provider.Targets() {
	// 	pattern := filepath.Join(working, path, "**", ext)
	// 	matches, err := zglob.Glob(pattern)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	files = append(files, matches...)
	// }

	grp := sync.WaitGroup{}

	var formatErrors *multierror.Error
	for _, filename := range files {
		grp.Add(1)
		go func(filename string) {
			defer grp.Done()

			fle, err := fls.Open(filename)
			if err != nil {
				formatErrors = multierror.Append(formatErrors, err)
				return
			}

			r, err := provider.Format(ctx, cfg, fle)
			if err != nil {
				formatErrors = multierror.Append(formatErrors, err)
				return
			}

			err = afero.WriteReader(fls, filename, r)
			if err != nil {
				formatErrors = multierror.Append(formatErrors, err)
				return
			}
		}(filename)
	}

	grp.Wait()

	return formatErrors.ErrorOrNil()
}
