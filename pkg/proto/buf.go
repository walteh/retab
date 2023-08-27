package proto

import (
	"context"
	"io"
	"io/fs"

	"github.com/bufbuild/buf/private/bufpkg/bufmodule"
	"github.com/bufbuild/buf/private/bufpkg/bufmodule/bufmoduleref"
	"github.com/spf13/afero"
)

// fileInfos, err := module.TargetFileInfos(ctx)
//
//	if err != nil {
//		return nil, err
//	}
//
// moduleFile, err := module.GetModuleFile(ctx, fileInfo.Path())
//
//	if err != nil {
//		return err
//	}
type Module struct {
	internal afero.Fs
	bufmodule.Module
}

var _ bufmodule.Module = (*Module)(nil)

func NewModule(ctx context.Context, internal afero.Fs) *Module {
	return &Module{
		internal: internal,
	}
}

func (m *Module) GetModuleFile(ctx context.Context, path string) (bufmodule.ModuleFile, error) {
	fle, err := m.internal.Open(path)
	if err != nil {
		return nil, err
	}
	return NewBufModuleFile(ctx, fle), nil
}

func (m *Module) TargetFileInfos(ctx context.Context) ([]bufmoduleref.FileInfo, error) {
	wlk, err := afero.ReadDir(m.internal, ".")
	if err != nil {
		return nil, err
	}

	fileInfos := make([]bufmoduleref.FileInfo, len(wlk))
	for i, fle := range wlk {
		fileInfos[i] = NewModuleFileInfo(fle)
	}

	return fileInfos, nil
}

type ModuleFile struct {
	internal afero.File
	io.Reader
	bufmodule.ModuleFile
}

var _ bufmodule.ModuleFile = (*ModuleFile)(nil)

func (m *ModuleFile) Path() string {
	return m.internal.Name()
}

func (m *ModuleFile) ExternalPath() string {
	return m.internal.Name()
}

func (m *ModuleFile) Close() error {
	return m.internal.Close()
}

func (m *ModuleFile) Read(p []byte) (n int, err error) {
	return m.internal.Read(p)
}

func NewBufModuleFile(ctx context.Context, fle afero.File) *ModuleFile {
	return &ModuleFile{
		internal: fle,
	}
}

type ModuleFileInfo struct {
	path fs.FileInfo
	bufmoduleref.FileInfo
}

func NewModuleFileInfo(path fs.FileInfo) *ModuleFileInfo {
	return &ModuleFileInfo{
		path: path,
	}
}

func (m *ModuleFileInfo) Path() string {
	return m.path.Name()
}
