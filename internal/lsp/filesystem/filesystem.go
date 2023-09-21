// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package filesystem

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"

	"github.com/spf13/afero"
	"github.com/walteh/retab/internal/lsp/document"
)

// Filesystem provides io/fs.FS compatible two-layer read-only filesystem
// with preferred source being DocumentStore and native OS FS acting as fallback.
//
// This allows for reading files in a directory while reflecting unsaved changes.
type Filesystem struct {
	osFs     afero.Fs
	docStore DocumentStore

	logger *log.Logger
}

type DocumentStore interface {
	GetDocument(document.Handle) (*document.Document, error)
	ListDocumentsInDir(document.DirHandle) ([]*document.Document, error)
}

func NewFilesystem(docStore DocumentStore) *Filesystem {
	return &Filesystem{
		osFs:     afero.NewOsFs(),
		docStore: docStore,
		logger:   log.New(io.Discard, "", 0),
	}
}

func (fs *Filesystem) SetLogger(logger *log.Logger) {
	fs.logger = logger
}

func (fs *Filesystem) ReadFile(name string) ([]byte, error) {
	doc, err := fs.docStore.GetDocument(document.HandleFromPath(name))
	if err != nil {
		if errors.Is(err, &document.DocumentNotFound{}) {
			return afero.ReadFile(fs.osFs, name)
		}
		return nil, err
	}

	return []byte(doc.Text), err
}

func (me *Filesystem) ReadDir(name string) ([]fs.DirEntry, error) {
	dirHandle := document.DirHandleFromPath(name)
	docList, err := me.docStore.ListDocumentsInDir(dirHandle)
	if err != nil {
		return nil, fmt.Errorf("doc FS: %w", err)
	}

	osList, err := afero.ReadDir(me.osFs, name)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("OS FS: %w", err)
	}

	list := documentsAsDirEntries(docList)
	for _, osEntry := range osList {
		if entryIsInList(list, fs.FileInfoToDirEntry(osEntry)) {
			continue
		}
		list = append(list, fs.FileInfoToDirEntry(osEntry))
	}

	return list, nil
}

func entryIsInList(list []fs.DirEntry, entry fs.DirEntry) bool {
	for _, di := range list {
		if di.Name() == entry.Name() {
			return true
		}
	}
	return false
}

func (fs *Filesystem) Open(name string) (afero.File, error) {
	doc, err := fs.docStore.GetDocument(document.HandleFromPath(name))
	if err != nil {
		if errors.Is(err, &document.DocumentNotFound{}) {
			return fs.osFs.Open(name)
		}
		return nil, err
	}

	return documentAsFile(doc)
}

func (fs *Filesystem) Stat(name string) (os.FileInfo, error) {
	doc, err := fs.docStore.GetDocument(document.HandleFromPath(name))
	if err != nil {
		if errors.Is(err, &document.DocumentNotFound{}) {
			return fs.osFs.Stat(name)
		}
		return nil, err
	}

	return documentAsFileInfo(doc), err
}
