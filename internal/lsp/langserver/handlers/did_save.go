// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	lsp "github.com/walteh/retab/gen/gopls"
)

func (svc *service) TextDocumentDidSave(ctx context.Context, params lsp.DidSaveTextDocumentParams) error {

	return nil
}
