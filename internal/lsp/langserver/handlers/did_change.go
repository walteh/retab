// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	"github.com/walteh/retab/gen/gopls"
	lsctx "github.com/walteh/retab/internal/lsp/context"
	"github.com/walteh/retab/internal/lsp/document"
	"github.com/walteh/retab/internal/lsp/lsp"
)

func (svc *service) TextDocumentDidChange(ctx context.Context, params gopls.DidChangeTextDocumentParams) error {
	p := gopls.DidChangeTextDocumentParams{
		TextDocument: gopls.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: gopls.TextDocumentIdentifier{
				URI: params.TextDocument.URI,
			},
			Version: params.TextDocument.Version,
		},
		ContentChanges: params.ContentChanges,
	}

	dh := lsp.HandleFromDocumentURI(p.TextDocument.URI)
	doc, err := svc.stateStore.DocumentStore.GetDocument(dh)
	if err != nil {
		return err
	}

	ctx = lsctx.WithLanguageId(ctx, doc.LanguageID)

	newVersion := int(p.TextDocument.Version)

	// Versions don't have to be consecutive, but they must be increasing
	if newVersion <= doc.Version {
		svc.logger.Printf("Old document version (%d) received, current version is %d. "+
			"Ignoring this update for %s. This is likely a client bug, please report it.",
			newVersion, doc.Version, p.TextDocument.URI)
		return nil
	}

	changes := lsp.DocumentChanges(params.ContentChanges)
	newText, err := document.ApplyChanges(doc.Text, changes)
	if err != nil {
		return err
	}
	err = svc.stateStore.DocumentStore.UpdateDocument(dh, newText, newVersion)
	if err != nil {
		return err
	}

	// jobIds, err := svc.indexer.DocumentChanged(ctx, dh.Dir)
	// if err != nil {
	// 	return err
	// }

	return svc.stateStore.JobStore.WaitForJobs(ctx)
}
