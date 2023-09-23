// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	gopls "github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/lsp/lsp"
)

func (svc *service) TextDocumentLink(ctx context.Context, params gopls.DocumentLinkParams) ([]gopls.DocumentLink, error) {
	cc, err := lsp.ClientCapabilities(ctx)
	if err != nil {
		return nil, err
	}

	filename := string(params.TextDocument.URI)

	// if doc.LanguageID != lsp.Retab.String() {
	// 	return nil, nil
	// }

	d, err := svc.decoderForDocument(ctx, filename)
	if err != nil {
		return nil, err
	}

	links, err := d.LinksInFile(filename)
	if err != nil {
		return nil, err
	}

	return lsp.Links(links, cc.TextDocument.DocumentLink), nil
}
