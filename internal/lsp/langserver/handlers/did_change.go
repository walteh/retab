// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"bytes"
	"context"

	"github.com/spf13/afero"
	"github.com/walteh/retab/gen/gopls"
	"github.com/walteh/retab/internal/lsp/document"
	"github.com/walteh/retab/internal/lsp/lsp"
)

func (svc *service) TextDocumentDidChange(ctx context.Context, params gopls.DidChangeTextDocumentParams) error {
	// p := gopls.DidChangeTextDocumentParams{
	// 	TextDocument: gopls.VersionedTextDocumentIdentifier{
	// 		TextDocumentIdentifier: gopls.TextDocumentIdentifier{
	// 			URI: params.TextDocument.URI,
	// 		},
	// 		Version: params.TextDocument.Version,
	// 	},
	// 	ContentChanges: params.ContentChanges,
	// }

	filename := string(params.TextDocument.URI)

	// ctx = lsctx.WithLanguageId(ctx, doc.LanguageID)

	// newVersion := int(p.TextDocument.Version)

	// // Versions don't have to be consecutive, but they must be increasing
	// if newVersion <= doc.Version {
	// 	svc.logger.Printf("Old document version (%d) received, current version is %d. "+
	// 		"Ignoring this update for %s. This is likely a client bug, please report it.",
	// 		newVersion, doc.Version, p.TextDocument.URI)
	// 	return nil
	// }

	text, err := afero.ReadFile(svc.fs, filename)
	if err != nil {
		return err
	}

	changes := lsp.DocumentChanges(params.ContentChanges)
	newText, err := document.ApplyChanges(text, changes)
	if err != nil {
		return err
	}

	// jobIds, err := svc.indexer.DocumentChanged(ctx, dh.Dir)
	// if err != nil {
	// 	return err
	// }

	return afero.WriteReader(svc.fs, filename, bytes.NewReader(newText))
}
