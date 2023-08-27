package git

import (
	"io/fs"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	"github.com/spf13/afero"
)

var _ billy.Filesystem = (*AferoBillyFs)(nil)

type AferoBillyFs struct {
	*afero.BasePathFs
}

func (me *AferoBillyFs) Fs() *afero.BasePathFs {
	return me.BasePathFs
}

func NewAferoBillyFs(internal afero.Fs, path string) *AferoBillyFs {
	return &AferoBillyFs{
		BasePathFs: afero.NewBasePathFs(internal, path).(*afero.BasePathFs),
	}
}

var _ billy.File = (*AferoBillyFile)(nil)

type AferoBillyFile struct {
	afero.File
}

func NewAferoBillyFile(internal afero.File) *AferoBillyFile {
	return &AferoBillyFile{
		internal,
	}
}

// Chroot implements billy.Filesystem.
func (me *AferoBillyFs) Chroot(path string) (billy.Filesystem, error) {
	return NewAferoBillyFs(me.Fs(), path), nil
}

// Create implements billy.Filesystem.
func (me *AferoBillyFs) Create(filename string) (billy.File, error) {
	fle, err := me.Fs().Create(filename)
	if err != nil {
		return nil, err
	}
	return NewAferoBillyFile(fle), nil
}

// Join implements billy.Filesystem.
func (me *AferoBillyFs) Join(elem ...string) string {
	return filepath.Join(elem...)
}

// Lstat implements billy.Filesystem.
func (me *AferoBillyFs) Lstat(filename string) (fs.FileInfo, error) {
	return me.Fs().Stat(filename)
}

// Open implements billy.Filesystem.
func (me *AferoBillyFs) Open(filename string) (billy.File, error) {
	fle, err := me.Fs().Open(filename)
	if err != nil {
		return nil, err
	}
	return NewAferoBillyFile(fle), nil
}

// OpenFile implements billy.Filesystem.
func (me *AferoBillyFs) OpenFile(filename string, flag int, perm fs.FileMode) (billy.File, error) {
	fle, err := me.Fs().OpenFile(filename, flag, perm)
	if err != nil {
		return nil, err
	}
	return NewAferoBillyFile(fle), nil
}

// ReadDir implements billy.Filesystem.
func (me *AferoBillyFs) ReadDir(path string) ([]fs.FileInfo, error) {
	return afero.ReadDir(me.Fs(), path)
}

// Readlink implements billy.Filesystem.
func (me *AferoBillyFs) Readlink(link string) (string, error) {
	p, err := me.Fs().ReadlinkIfPossible(link)
	if err != nil {
		return "", err
	} else {
		return p, nil
	}
}

// Root implements billy.Filesystem.
func (me *AferoBillyFs) Root() string {
	p, err := me.Fs().RealPath("/")
	if err != nil {
		return "/"
	} else {
		return p
	}
}

// Symlink implements billy.Filesystem.
func (me *AferoBillyFs) Symlink(target string, link string) error {
	return me.Fs().SymlinkIfPossible(target, link)
}

// TempFile implements billy.Filesystem.
func (me *AferoBillyFs) TempFile(dir string, prefix string) (billy.File, error) {
	fle, err := afero.TempFile(me.Fs(), dir, prefix)
	if err != nil {
		return nil, err
	}

	return NewAferoBillyFile(fle), nil
}

// Lock implements billy.File.
func (me *AferoBillyFile) Lock() error {
	return nil
}

// Unlock implements billy.File.
func (me *AferoBillyFile) Unlock() error {
	return nil
}
