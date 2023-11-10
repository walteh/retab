package file

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/spf13/afero"
)

func CopyFile(_ context.Context, src, dest afero.File) error {
	// Rewind the source file in case it has been read before
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// Use io.Copy to copy from src to dest
	if _, err := io.Copy(dest, src); err != nil {
		return err
	}

	return nil
}

func CopyDirectory(ctx context.Context, srcFs afero.Fs, destFs afero.Fs, inputpath string) error {
	inputpath = filepath.Join(inputpath, "/")
	zerolog.Ctx(ctx).Debug().Str("path", inputpath).Msg("copying directory")
	return afero.Walk(srcFs, inputpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		zerolog.Ctx(ctx).Debug().Str("path", path).Msg("copying file")

		// Create file in destination
		srcFile, err := srcFs.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// Get relative path
		// relPath, err := filepath.Rel("/", path)
		// if err != nil {
		// 	return err
		// }
		// destPath := filepath.Join("/", relPath)

		if info.IsDir() {
			// zerolog.Ctx(ctx).Debug().Str("path", path).Msg("creating directory")
			// // Create directory in destination
			// err := destFs.MkdirAll(path, info.Mode())
			// if err != nil {
			// 	return err
			// }

			// return CopyDirectory(ctx, srcFs, destFs, path)
			return nil
		}

		zerolog.Ctx(ctx).Debug().Str("path", path).Msg("opening file in destination")

		zerolog.Ctx(ctx).Debug().Str("path", path).Msg("creating file in destination")

		destFile, err := destFs.Create(path)
		if err != nil {
			return err
		}
		defer destFile.Close()

		err = CopyFile(ctx, srcFile, destFile)
		if err != nil {
			return err
		}

		zerolog.Ctx(ctx).Debug().Str("path", path).Msg("setting file permissions")

		// Optional: Set file permissions
		return destFs.Chmod(path, info.Mode())
	})
}
