package file

import (
	"context"
	"embed"
	"io"
	"io/fs"
	"time"

	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/afero"
)

type rawCopyableReadOnlyEmbedFs struct {
	embedfs embed.FS
}

var _ afero.Fs = (*rawCopyableReadOnlyEmbedFs)(nil)

func NewEmbedFs(ctx context.Context, emfs embed.FS, dir string) (afero.Fs, error) {

	rofsem := newRawCopyableReadOnlyEmbedFs(emfs)

	outfs := afero.NewMemMapFs()

	err := CopyDirectory(ctx, rofsem, outfs, dir)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("failed to copy embed fs")
		return nil, err
	}

	return afero.NewBasePathFs(outfs, dir), nil
}

func newRawCopyableReadOnlyEmbedFs(emfs embed.FS) *rawCopyableReadOnlyEmbedFs {
	return &rawCopyableReadOnlyEmbedFs{
		embedfs: emfs,
	}
}

func (e *rawCopyableReadOnlyEmbedFs) Open(name string) (afero.File, error) {
	file, err := e.embedfs.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stt, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if stt.IsDir() {
		outer, err := e.embedfs.ReadDir(name)
		if err != nil {
			return nil, err
		}

		return NewReadOnlyEmbedDir(file, outer)

	}

	fle, err := afero.NewMemMapFs().Create(name)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(fle, file)
	if err != nil {
		return nil, err
	}

	return fle, nil
}

func (e *rawCopyableReadOnlyEmbedFs) ReadFile(name string) ([]byte, error) {
	return e.embedfs.ReadFile(name)
}

func (e *rawCopyableReadOnlyEmbedFs) ReadDir(name string) ([]fs.FileInfo, error) {
	dir, err := e.embedfs.ReadDir(name)
	if err != nil {
		return nil, err
	}

	files := make([]fs.FileInfo, 0, len(dir))

	for i, file := range dir {
		fi, err := file.Info()
		if err != nil {
			return nil, err
		}

		files[i] = fi
	}

	return files, nil
}

func (e *rawCopyableReadOnlyEmbedFs) Stat(name string) (fs.FileInfo, error) {
	file, err := e.embedfs.Open(name)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	return file.Stat()
}

func (e *rawCopyableReadOnlyEmbedFs) OpenFile(name string, _ int, _ fs.FileMode) (afero.File, error) {
	return e.Open(name)
}

// Chmod implements afero.Fs.
func (*rawCopyableReadOnlyEmbedFs) Chmod(_ string, _ fs.FileMode) error {
	return fs.ErrInvalid
}

// Chown implements afero.Fs.
func (*rawCopyableReadOnlyEmbedFs) Chown(_ string, _ int, _ int) error {
	return fs.ErrInvalid
}

// Chtimes implements afero.Fs.
func (*rawCopyableReadOnlyEmbedFs) Chtimes(_ string, _ time.Time, _ time.Time) error {
	return fs.ErrInvalid
}

// Create implements afero.Fs.
func (*rawCopyableReadOnlyEmbedFs) Create(_ string) (afero.File, error) {
	return nil, fs.ErrInvalid
}

// Mkdir implements afero.Fs.
func (*rawCopyableReadOnlyEmbedFs) Mkdir(_ string, _ fs.FileMode) error {
	return fs.ErrInvalid
}

// MkdirAll implements afero.Fs.
func (*rawCopyableReadOnlyEmbedFs) MkdirAll(_ string, _ fs.FileMode) error {
	return fs.ErrInvalid
}

// Name implements afero.Fs.
func (*rawCopyableReadOnlyEmbedFs) Name() string {
	return "readonlyembedfs"
}

// Remove implements afero.Fs.
func (*rawCopyableReadOnlyEmbedFs) Remove(_ string) error {
	return fs.ErrInvalid
}

// RemoveAll implements afero.Fs.
func (*rawCopyableReadOnlyEmbedFs) RemoveAll(_ string) error {
	return fs.ErrInvalid
}

// Rename implements afero.Fs.
func (*rawCopyableReadOnlyEmbedFs) Rename(_ string, _ string) error {
	return fs.ErrInvalid
}

var _ afero.File = (*ReadOnlyEmbedDir)(nil)

type ReadOnlyEmbedDir struct {
	fsfile     fs.File
	direntries []fs.DirEntry
}

func NewReadOnlyEmbedDir(fsfile fs.File, direntries []fs.DirEntry) (*ReadOnlyEmbedDir, error) {

	return &ReadOnlyEmbedDir{
		fsfile:     fsfile,
		direntries: direntries,
	}, nil
}

func (r *ReadOnlyEmbedDir) Close() error {
	return fs.ErrInvalid
}

func (r *ReadOnlyEmbedDir) Name() string {
	return "readonlyembeddir"
}

func (r *ReadOnlyEmbedDir) Read(_ []byte) (n int, err error) {
	return 0, fs.ErrInvalid
}

func (r *ReadOnlyEmbedDir) ReadAt(_ []byte, _ int64) (n int, err error) {
	return 0, fs.ErrInvalid
}

func (r *ReadOnlyEmbedDir) Readdir(_ int) ([]fs.FileInfo, error) {
	fileinfos := make([]fs.FileInfo, 0, len(r.direntries))
	for i, entry := range r.direntries {
		fi, err := entry.Info()
		if err != nil {
			return nil, err
		}

		fileinfos[i] = fi
	}

	return fileinfos, nil
}

func (r *ReadOnlyEmbedDir) Readdirnames(_ int) ([]string, error) {
	str := make([]string, len(r.direntries))
	for i, entry := range r.direntries {
		str[i] = entry.Name()
	}
	return str, nil
}

func (r *ReadOnlyEmbedDir) Seek(_ int64, _ int) (int64, error) {
	return 0, fs.ErrInvalid
}

func (r *ReadOnlyEmbedDir) Stat() (os.FileInfo, error) {
	return r.fsfile.Stat()
}

func (r *ReadOnlyEmbedDir) Sync() error {
	return fs.ErrInvalid
}

func (r *ReadOnlyEmbedDir) Truncate(_ int64) error {
	return fs.ErrInvalid
}

func (r *ReadOnlyEmbedDir) Write(_ []byte) (n int, err error) {
	return 0, fs.ErrInvalid
}

func (r *ReadOnlyEmbedDir) WriteAt(_ []byte, _ int64) (n int, err error) {
	return 0, fs.ErrInvalid
}

func (r *ReadOnlyEmbedDir) WriteString(_ string) (ret int, err error) {
	return 0, fs.ErrInvalid
}
