// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package filesystem

import (
	"io/fs"

	"github.com/spf13/afero"
	"github.com/walteh/retab/internal/lsp/document"
)

func documentAsFile(doc *document.Document) (afero.File, error) {

	fle, err := afero.NewMemMapFs().Open(doc.Filename)
	if err != nil {
		return nil, err
	}

	_, err = fle.Write(doc.Text)
	if err != nil {
		return nil, err
	}

	return fle, nil
}

func documentAsFileInfo(doc *document.Document) fs.FileInfo {
	return inMemFileInfo{
		name:    doc.Filename,
		size:    len(doc.Text),
		modTime: doc.ModTime,
		mode:    0o755,
		isDir:   false,
	}
}

func documentsAsDirEntries(docs []*document.Document) []fs.DirEntry {
	entries := make([]fs.DirEntry, len(docs))

	for i, doc := range docs {
		entries[i] = documentAsDirEntry(doc)
	}

	return entries
}

func documentAsDirEntry(doc *document.Document) fs.DirEntry {
	return fs.FileInfoToDirEntry(documentAsFileInfo(doc))
}
