// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	gopls "github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/lsp/lsp"
)

func (svc *service) TextDocumentComplete(ctx context.Context, params gopls.CompletionParams) (gopls.CompletionList, error) {
	var list gopls.CompletionList

	cc, err := lsp.ClientCapabilities(ctx)
	if err != nil {
		return list, err
	}

	filename := string(params.TextDocument.URI)

	d, err := svc.decoderForDocument(ctx, filename)
	if err != nil {
		return list, err
	}

	pos, err := lsp.HCLPositionFromLspPosition(params.TextDocumentPositionParams.Position, svc.fs, filename)
	if err != nil {
		return list, err
	}

	svc.logger.Printf("Looking for candidates at %q -> %#v", filename, pos)
	candidates, err := d.CandidatesAtPos(ctx, filename, pos)
	svc.logger.Printf("received candidates: %#v", candidates)
	return lsp.ToCompletionList(candidates, cc.TextDocument), err
}
