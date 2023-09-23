// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	gopls "github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/lsp/lsp"
)

func (svc *service) TextDocumentHover(ctx context.Context, params gopls.TextDocumentPositionParams) (*gopls.Hover, error) {
	cc, err := lsp.ClientCapabilities(ctx)
	if err != nil {
		return nil, err
	}

	filename := string(params.TextDocument.URI)

	d, err := svc.decoderForDocument(ctx, filename)
	if err != nil {
		return nil, err
	}

	pos, err := lsp.HCLPositionFromLspPosition(params.Position, svc.fs, filename)
	if err != nil {
		return nil, err
	}

	svc.logger.Printf("Looking for hover data at %q -> %#v", filename, pos)
	hoverData, err := d.HoverAtPos(ctx, filename, pos)
	svc.logger.Printf("received hover data: %#v", hoverData)
	if err != nil {
		return nil, err
	}

	return lsp.HoverData(hoverData, cc.TextDocument), nil
}
