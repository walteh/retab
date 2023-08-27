package format

import (
	"context"
	"io"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/mattn/go-zglob"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/walteh/tftab/pkg/configuration"
	"github.com/walteh/tftab/pkg/configuration/editorconfig"
)

type Provider interface {
	Format(ctx context.Context, cfg configuration.Provider, reader io.Reader) (io.Reader, error)
	Targets() []string
}

func Format(ctx context.Context, provider Provider, fs afero.Fs, path string, working string) error {

	isDir, err := afero.IsDir(fs, path)
	if err != nil {
		return err
	}

	// handle when option specifies a particular file
	if !isDir {
		cfg, err := editorconfig.NewEditorConfigConfigurationProvider(ctx, path)
		if err != nil {
			return err
		}

		if !filepath.IsAbs(path) {
			path = filepath.Join(working, path)
		}

		zerolog.Ctx(ctx).Debug().Msgf("Formatting hcl file at: %s.", path)

		fle, err := fs.Open(path)
		if err != nil {
			return err
		}

		r, err := provider.Format(ctx, cfg, fle)
		if err != nil {
			return err
		}

		err = afero.WriteReader(fs, path, r)
		if err != nil {
			return err
		}

		zerolog.Ctx(ctx).Debug().Msgf("Formatted file at: %s.", path)

		return nil

	}

	zerolog.Ctx(ctx).Debug().Msgf("Formatting hcl files from the directory tree %s %s", working, path)

	// zglob normalizes paths to "/"
	var files []string

	for _, ext := range provider.Targets() {
		pattern := filepath.Join(working, path, "**", ext)
		matches, err := zglob.Glob(pattern)
		if err != nil {
			return err
		}
		files = append(files, matches...)
	}

	var formatErrors *multierror.Error
	for _, filename := range files {
		cfg, err := editorconfig.NewEditorConfigConfigurationProvider(ctx, filename)
		if err != nil {
			formatErrors = multierror.Append(formatErrors, err)
			continue
		}
		fle, err := fs.Open(path)
		if err != nil {
			formatErrors = multierror.Append(formatErrors, err)
			continue
		}

		r, err := provider.Format(ctx, cfg, fle)
		if err != nil {
			formatErrors = multierror.Append(formatErrors, err)
			continue
		}

		err = afero.WriteReader(fs, path, r)
		if err != nil {
			formatErrors = multierror.Append(formatErrors, err)
			continue
		}
	}

	return formatErrors.ErrorOrNil()
}
