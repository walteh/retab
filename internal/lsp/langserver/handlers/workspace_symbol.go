// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	gopls "github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/lsp/lsp"
)

func (svc *service) WorkspaceSymbol(ctx context.Context, params gopls.WorkspaceSymbolParams) ([]gopls.SymbolInformation, error) {
	cc, err := lsp.ClientCapabilities(ctx)
	if err != nil {
		return nil, err
	}

	symbols, err := svc.decoder.Symbols(ctx, params.Query)
	if err != nil {
		return nil, err
	}

	return lsp.WorkspaceSymbols(symbols, cc.Workspace.Symbol), nil
}
