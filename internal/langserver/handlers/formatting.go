// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
	lsp "github.com/walteh/retab/gen/gopls"
	"github.com/walteh/retab/internal/document"
	"github.com/walteh/retab/internal/hcl"
	ilsp "github.com/walteh/retab/internal/lsp"
	"github.com/walteh/retab/internal/source"
)

func (svc *service) TextDocumentFormatting(ctx context.Context, params lsp.DocumentFormattingParams) ([]lsp.TextEdit, error) {
	var edits []lsp.TextEdit

	dh := ilsp.HandleFromDocumentURI(params.TextDocument.URI)

	text, err := afero.ReadFile(svc.fs, dh.Filename)
	if err != nil {
		return edits, err
	}

	doc := &document.Document{
		Dir:        dh.Dir,
		Filename:   dh.Filename,
		ModTime:    time.Now(),
		LanguageID: filepath.Ext(dh.Filename),
		Version:    1,
		Text:       text,
		Lines:      source.MakeSourceLines(dh.Filename, text),
	}

	edits, err = svc.formatDocument(ctx, doc.Text, dh)
	if err != nil {
		return edits, err
	}

	return edits, nil
}

func (svc *service) formatDocument(ctx context.Context, original []byte, dh document.Handle) ([]lsp.TextEdit, error) {
	// var edits []lsp.TextEdit

	svc.logger.Printf("formatting document via")

	startTime := time.Now()
	formatted := []byte{}
	svc.logger.Printf("Finished 'terraform fmt' in %s", time.Now().Sub(startTime))

	changes := hcl.Diff(dh, original, formatted)

	return ilsp.TextEditsFromDocumentChanges(changes), nil
}
