// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	"github.com/walteh/retab/gen/gopls"
	"github.com/walteh/retab/internal/lsp/lsp"
)

func (svc *service) TextDocumentComplete(ctx context.Context, params gopls.CompletionParams) (gopls.CompletionList, error) {
	var list gopls.CompletionList

	cc, err := lsp.ClientCapabilities(ctx)
	if err != nil {
		return list, err
	}

	dh := lsp.HandleFromDocumentURI(params.TextDocument.URI)
	doc, err := svc.stateStore.DocumentStore.GetDocument(dh)
	if err != nil {
		return list, err
	}

	d, err := svc.decoderForDocument(ctx, doc)
	if err != nil {
		return list, err
	}

	pos, err := lsp.HCLPositionFromLspPosition(params.TextDocumentPositionParams.Position, doc)
	if err != nil {
		return list, err
	}

	svc.logger.Printf("Looking for candidates at %q -> %#v", doc.Filename, pos)
	candidates, err := d.CandidatesAtPos(ctx, doc.Filename, pos)
	svc.logger.Printf("received candidates: %#v", candidates)
	return lsp.ToCompletionList(candidates, cc.TextDocument), err
}
