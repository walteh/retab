// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	"github.com/walteh/retab/gen/gopls"
	"github.com/walteh/retab/internal/lsp/lsp"
)

func (svc *service) TextDocumentSymbol(ctx context.Context, params gopls.DocumentSymbolParams) ([]gopls.DocumentSymbol, error) {
	var symbols []gopls.DocumentSymbol

	cc, err := lsp.ClientCapabilities(ctx)
	if err != nil {
		return symbols, err
	}

	filename := string(params.TextDocument.URI)

	d, err := svc.decoderForDocument(ctx, filename)
	if err != nil {
		return symbols, err
	}

	sbs, err := d.SymbolsInFile(filename)
	if err != nil {
		return symbols, err
	}

	return lsp.DocumentSymbols(sbs, cc.TextDocument.DocumentSymbol), nil
}
