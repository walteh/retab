// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	lsp "github.com/walteh/retab/gen/gopls"
	lsctx "github.com/walteh/retab/internal/context"
	"github.com/walteh/retab/internal/langserver/cmd"
	"github.com/walteh/retab/internal/langserver/handlers/command"
	ilsp "github.com/walteh/retab/internal/lsp"
)

func (svc *service) TextDocumentDidSave(ctx context.Context, params lsp.DidSaveTextDocumentParams) error {
	expFeatures, err := lsctx.ExperimentalFeatures(ctx)
	if err != nil {
		return err
	}
	if !expFeatures.ValidateOnSave {
		return nil
	}

	dh := ilsp.HandleFromDocumentURI(params.TextDocument.URI)

	cmdHandler := &command.CmdHandler{
		StateStore: svc.stateStore,
	}
	_, err = cmdHandler.TerraformValidateHandler(ctx, cmd.CommandArgs{
		"uri": dh.Dir.URI,
	})

	return err
}
