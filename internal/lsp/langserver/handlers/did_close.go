// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	gopls "github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/lsp/lsp"
)

func (svc *service) TextDocumentDidClose(ctx context.Context, params gopls.DidCloseTextDocumentParams) error {
	dh := lsp.HandleFromDocumentURI(params.TextDocument.URI)
	return svc.stateStore.DocumentStore.CloseDocument(dh)
}
