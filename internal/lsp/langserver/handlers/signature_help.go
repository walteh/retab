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

	filename := string(params.TextDocument.URI)

	d, err := svc.decoderForDocument(ctx, filename)
	if err != nil {
		return nil, err
	}

	pos, err := lsp.HCLPositionFromLspPosition(params.Position, svc.fs, filename)
	if err != nil {
		return nil, err
	}

	sig, err := d.SignatureAtPos(filename, pos)
	if err != nil {
		return nil, err
	}

	return lsp.ToSignatureHelp(sig), nil
}
