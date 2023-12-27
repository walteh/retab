package resolvers

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/walteh/snake"
	"github.com/walteh/terrors"
)

func FSRunner() snake.Runner {
	return snake.GenRunResolver_In01_Out03(&FSResolver{})
}

type FSResolver struct {
	Dir  string `usage:"the directory to run in"`
	File string `usage:"the file to read the configuration from"`
}

func (me *FSResolver) Run(ctx context.Context) (afero.Fs, afero.File, error) {
	res := afero.NewOsFs()
	if !filepath.IsAbs(me.File) {
		if me.Dir == "" {
			wrking, err := os.Getwd()
			if err != nil {
				return nil, nil, err
			}
			res = afero.NewBasePathFs(res, wrking)
		} else {
			res = afero.NewBasePathFs(res, me.Dir)
		}
	} else {
		zerolog.Ctx(ctx).Warn().Msg("absolute path given for directory, ignoring")
	}
	path := me.File

	if path == "" {
		path = "."
	}

	fle, err := res.Open(path)
	if err != nil {
		return res, nil, terrors.Wrap(err, "failed to open file")
	}

	return res, fle, nil
}

func GetFileOrGlobDir(ctx context.Context, fs afero.Fs, fle afero.File, glob string) ([]string, error) {
	isDir, err := afero.IsDir(fs, fle.Name())
	if err != nil {
		return nil, err
	}

	fles := []string{}

	if isDir {
		flesd, err := afero.Glob(fs, glob)
		if err != nil {
			return nil, err
		}
		fles = append(fles, flesd...)
	} else {
		fles = append(fles, fle.Name())
	}

	return fles, nil
}

func ForAllFilesAtSameTime(ctx context.Context, fls afero.Fs, files []string, cb func(ctx context.Context, fle afero.File) (io.Reader, error)) error {

	grp := sync.WaitGroup{}

	var formatErrors *multierror.Error
	for _, filename := range files {
		grp.Add(1)
		go func(filename string) {
			defer grp.Done()

			datafunc := func(e *zerolog.Event) *zerolog.Event {
				return e.Str("path", filename)
			}

			fle, err := fls.Open(filename)
			if err != nil {
				formatErrors = multierror.Append(formatErrors, terrors.Wrap(err, "failed to open file").Event(datafunc))
				return
			}

			defer fle.Close()

			r, err := cb(ctx, fle)
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
