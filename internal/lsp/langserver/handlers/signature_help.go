// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	gopls "github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/lsp/lsp"
)

func (svc *service) SignatureHelp(ctx context.Context, params gopls.SignatureHelpParams) (*gopls.SignatureHelp, error) {
	_, err := lsp.ClientCapabilities(ctx)
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

	sig, err := d.SignatureAtPos(doc.Filename, pos)
	if err != nil {
		return nil, err
	}

	return lsp.ToSignatureHelp(sig), nil
}
