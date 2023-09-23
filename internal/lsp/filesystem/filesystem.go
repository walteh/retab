// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package filesystem

import (
	"io"
	"io/fs"
	"log"

	"github.com/spf13/afero"
	"github.com/walteh/retab/internal/lsp/document"
)

// Filesystem provides io/fs.FS compatible two-layer read-only filesystem
// with preferred source being DocumentStore and native OS FS acting as fallback.
//
// This allows for reading files in a directory while reflecting unsaved changes.
type Filesystem struct {
	afero.Fs

	logger *log.Logger
}

type DocumentStore interface {
	GetDocument(document.Handle) (*document.Document, error)
	ListDocumentsInDir(document.DirHandle) ([]*document.Document, error)
}

func NewFilesystem(fls afero.Fs) *Filesystem {
	return &Filesystem{
		Fs:     fls,
		logger: log.New(io.Discard, "", 0),
	}
}

func (fs *Filesystem) SetLogger(logger *log.Logger) {
	fs.logger = logger
}

// func (fs *Filesystem) ReadFile(name string) ([]byte, error) {
// 	return afero.ReadFile(fs, name)
// }

// func (me *Filesystem) ReadDir(name string) ([]fs.DirEntry, error) {

// 	afero.Re
// 	dirHandle := document.DirHandleFromPath(name)
// 	docList, err := me.docStore.ListDocumentsInDir(dirHandle)
// 	if err != nil {
// 		return nil, fmt.Errorf("doc FS: %w", err)
// 	}

// 	osList, err := afero.ReadDir(me.fls, name)
// 	if err != nil && !os.IsNotExist(err) {
// 		return nil, fmt.Errorf("OS FS: %w", err)
// 	}

// 	list := documentsAsDirEntries(docList)
// 	for _, osEntry := range osList {
// 		if entryIsInList(list, fs.FileInfoToDirEntry(osEntry)) {
// 			continue
// 		}
// 		list = append(list, fs.FileInfoToDirEntry(osEntry))
// 	}

// 	return list, nil
// }

func entryIsInList(list []fs.DirEntry, entry fs.DirEntry) bool {
	for _, di := range list {
		if di.Name() == entry.Name() {
			return true
		}
	}
	return false
}

// func (fs *Filesystem) Open(name string) (afero.File, error) {
// 	doc, err := fs.docStore.GetDocument(document.HandleFromPath(name))
// 	if err != nil {
// 		if errors.Is(err, &document.DocumentNotFound{}) {
// 			return fs.fls.Open(name)
// 		}
// 		return nil, err
// 	}

// 	return documentAsFile(doc)
// }

// func (fs *Filesystem) Stat(name string) (os.FileInfo, error) {
// 	doc, err := fs.docStore.GetDocument(document.HandleFromPath(name))
// 	if err != nil {
// 		if errors.Is(err, &document.DocumentNotFound{}) {
// 			return fs.fls.Stat(name)
// 		}
// 		return nil, err
// 	}

// 	return documentAsFileInfo(doc), err
// }
