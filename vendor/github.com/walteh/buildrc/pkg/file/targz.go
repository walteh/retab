package file

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/afero"
)

func wrap(ctx context.Context, err error) error {
	zerolog.Ctx(ctx).Error().Err(err).CallerSkipFrame(1).Msg("")
	return err
}
func Targz(ctx context.Context, fs afero.Fs, pth string) (afero.File, error) {

	wrk, err := fs.Create(pth + ".tar.gz")
	if err != nil {
		return nil, wrap(ctx, err)
	}

	writer, err := gzip.NewWriterLevel(wrk, gzip.BestCompression)
	if err != nil {
		return nil, wrap(ctx, err)
	}
	defer writer.Close()

	tw := tar.NewWriter(writer)
	defer tw.Close()

	fle, err := fs.Open(pth)
	if err != nil {
		return nil, wrap(ctx, err)
	}

	defer fle.Close()

	if err := addFilesToTar(ctx, fs, tw, fle); err != nil {
		return nil, wrap(ctx, err)
	}

	return wrk, nil
}

func addFilesToTar(ctx context.Context, fls afero.Fs, tw *tar.Writer, file afero.File) error {

	stats, err := file.Stat()
	if err != nil {
		return wrap(ctx, err)
	}

	if stats.IsDir() {

		infos, err := file.Readdirnames(-1)
		if err != nil {
			return wrap(ctx, err)
		}

		for _, info := range infos {
			zerolog.Ctx(ctx).Trace().Str("item", info).Str("dir", file.Name()).Msg("opening item in dir")
			fle, err := fls.Open(filepath.Join(file.Name(), info))
			if err != nil {
				return wrap(ctx, err)
			}

			defer fle.Close()

			if err := addFilesToTar(ctx, fls, tw, fle); err != nil {
				return wrap(ctx, err)
			}
		}

		hdr := &tar.Header{
			Name:     file.Name(), // Set the name to the relative path
			Mode:     int64(stats.Mode()),
			ModTime:  stats.ModTime(),
			Format:   tar.FormatGNU,
			Typeflag: tar.TypeDir,
		}

		if err := tw.WriteHeader(hdr); err != nil {
			return wrap(ctx, err)
		}

		zerolog.Ctx(ctx).Trace().Str("path", file.Name()).Msg("added dir to tar")
		return nil
	}

	body, err := io.ReadAll(file)
	if err != nil {
		return wrap(ctx, err)
	}

	hdr := &tar.Header{
		Name:     file.Name(), // Set the name to the relative path
		Mode:     int64(stats.Mode()),
		Size:     int64(len(body)),
		ModTime:  stats.ModTime(),
		Format:   tar.FormatGNU,
		Typeflag: tar.TypeReg,
	}

	if err := tw.WriteHeader(hdr); err != nil {
		return wrap(ctx, err)
	}

	if _, err := tw.Write(body); err != nil {
		return wrap(ctx, err)
	}

	zerolog.Ctx(ctx).Trace().Str("path", file.Name()).Msg("added file to tar")

	return nil
}

func UntargzFile(ctx context.Context, fle afero.File) (afero.File, error) {
	fls := afero.NewMemMapFs()
	f, err := afero.TempFile(fls, "", "test")
	if err != nil {
		return nil, wrap(ctx, err)
	}

	_, err = io.Copy(f, fle)
	if err != nil {
		return nil, wrap(ctx, err)
	}

	return Untargz(ctx, fls, f.Name())
}

func Untargz(ctx context.Context, fs afero.Fs, pth string) (afero.File, error) {

	if !strings.HasSuffix(pth, ".tar.gz") && !strings.HasSuffix(pth, ".tgz") {
		return nil, wrap(ctx, errors.New("file is not a tar.gz or tgz"))
	}

	fle, err := fs.Open(pth)
	if err != nil {
		return nil, wrap(ctx, err)
	}
	defer fle.Close()

	dest := strings.TrimSuffix(pth, ".tar.gz")
	dest = strings.TrimSuffix(dest, ".tgz")

	gr, err := gzip.NewReader(fle)
	if err != nil {
		return nil, wrap(ctx, err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for i := 0; ; i++ {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, wrap(ctx, err)
		}

		destPath := filepath.Join(dest, hdr.Name) // Update the destination directory as needed
		if hdr.Typeflag == tar.TypeDir {
			if err := fs.MkdirAll(destPath, 0755); err != nil {

				return nil, wrap(ctx, err)
			}
			zerolog.Ctx(ctx).Trace().Str("path", destPath).Msg("created directory from tar")
			continue
		} else if i == 0 && hdr.Name == dest {
			// if it is the first and only file, we want to extract it to the same directory with the original name
			destPath = dest
		}

		destFile, err := fs.Create(destPath)
		if err != nil {
			return nil, wrap(ctx, err)
		}

		_, err = io.Copy(destFile, tr)
		if err != nil {
			return nil, wrap(ctx, err)
		}

		if err := destFile.Close(); err != nil {
			return nil, wrap(ctx, err)
		}

		zerolog.Ctx(ctx).Trace().Str("path", destPath).Msg("extracted file from tar")

	}

	dst, err := fs.Open(dest)
	if err != nil {
		return nil, wrap(ctx, err)
	}

	return dst, nil
}

// func Untargz(ctx context.Context, fls afero.Fs, pth string) (afero.File, error) {

// 	fle, err := fls.Open(pth)
// 	if err != nil {
// 		return nil, wrap(ctx, err,)
// 	}
// 	defer fle.Close()

// 	gr, err := gzip.NewReader(fle)
// 	if err != nil {
// 		return nil, wrap(ctx, err,)
// 	}
// 	defer gr.Close()

// 	tr := tar.NewReader(gr)

// 	// // Assuming you want to extract to the same directory with the original name
// 	// destPath := strings.TrimSuffix(fle.Name(), ".tar.gz")
// 	// destFile, err := fls.Create(destPath)
// 	// if err != nil {
// 	// 	return nil, wrap(ctx, err,)
// 	// }

// 	// Iterate through the files in the tar archive
// 	for {
// 		hdr, err := tr.Next()
// 		if err == io.EOF {
// 			break // Reached end of archive
// 		}
// 		if err != nil {
// 			return nil, wrap(ctx, err,)
// 		}

// 		// Create destination file based on header name
// 		destPath := filepath.Join("destination_directory", hdr.Name) // Update the destination directory as needed
// 		destFile, err := fls.Create(destPath)
// 		if err != nil {
// 			return nil, wrap(ctx, err,)
// 		}

// 		// Copy content to destination file
// 		_, err = io.Copy(destFile, tr)
// 		if err != nil {
// 			return nil, wrap(ctx, err,)
// 		}

// 		// Close destination file
// 		if err := destFile.Close(); err != nil {
// 			return nil, wrap(ctx, err,)
// 		}
// 	}

// 	return destFile, nil
// }
