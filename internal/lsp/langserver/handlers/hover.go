// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	"github.com/walteh/retab/gen/gopls"
	"github.com/walteh/retab/internal/lsp/lsp"
)

func (svc *service) TextDocumentHover(ctx context.Context, params gopls.TextDocumentPositionParams) (*gopls.Hover, error) {
	cc, err := lsp.ClientCapabilities(ctx)
	if err != nil {
		return nil, err
	}

	dh := lsp.HandleFromDocumentURI(params.TextDocument.URI)
	doc, err := svc.stateStore.DocumentStore.GetDocument(dh)
	if err != nil {
		return nil, err
	}

	d, err := svc.decoderForDocument(ctx, doc)
	if err != nil {
		return nil, err
	}

	pos, err := lsp.HCLPositionFromLspPosition(params.Position, doc)
	if err != nil {
		return nil, err
	}

	svc.logger.Printf("Looking for hover data at %q -> %#v", doc.Filename, pos)
	hoverData, err := d.HoverAtPos(ctx, doc.Filename, pos)
	svc.logger.Printf("received hover data: %#v", hoverData)
	if err != nil {
		return nil, err
	}

	return lsp.HoverData(hoverData, cc.TextDocument), nil
}
