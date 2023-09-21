// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	"github.com/creachadair/jrpc2"
	"github.com/walteh/retab/gen/gopls"
	"github.com/walteh/retab/internal/lsp/lsp"
)

func (svc *service) TextDocumentSemanticTokensFull(ctx context.Context, params gopls.SemanticTokensParams) (gopls.SemanticTokens, error) {
	tks := gopls.SemanticTokens{}

	cc, err := lsp.ClientCapabilities(ctx)
	if err != nil {
		return tks, err
	}

	caps := lsp.SemanticTokensClientCapabilities{
		SemanticTokensClientCapabilities: cc.TextDocument.SemanticTokens,
	}
	if !caps.FullRequest() {
		// This would indicate a buggy client which sent a request
		// it didn't claim to support, so we just strictly follow
		// the protocol here and avoid serving buggy clients.
		svc.logger.Printf("semantic tokens full request support not announced by client")
		return tks, jrpc2.MethodNotFound.Err()
	}

	dh := lsp.HandleFromDocumentURI(params.TextDocument.URI)
	doc, err := svc.stateStore.DocumentStore.GetDocument(dh)
	if err != nil {
		return tks, err
	}

	d, err := svc.decoderForDocument(ctx, doc)
	if err != nil {
		return tks, err
	}

	tokens, err := d.SemanticTokensInFile(ctx, doc.Filename)
	if err != nil {
		return tks, err
	}

	te := &lsp.TokenEncoder{
		Lines:      doc.Lines,
		Tokens:     tokens,
		ClientCaps: cc.TextDocument.SemanticTokens,
	}
	tks.Data = te.Encode()

	return tks, nil
}
