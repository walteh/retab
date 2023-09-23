// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"
	"fmt"

	"github.com/creachadair/jrpc2"
	gopls "github.com/walteh/retab/gen/gopls/protocol"
	lsctx "github.com/walteh/retab/internal/lsp/context"
	"github.com/walteh/retab/internal/lsp/langserver/cmd"
)

func cmdHandlers(svc *service) cmd.Handlers {
	// cmdHandler := &command.CmdHandler{
	// 	StateStore: svc.stateStore,
	// 	Logger:     svc.logger,
	// }
	return cmd.Handlers{}
}

func (svc *service) WorkspaceExecuteCommand(ctx context.Context, params gopls.ExecuteCommandParams) (interface{}, error) {
	if params.Command == "editor.action.triggerSuggest" {
		// If this was actually received by the server, it means the client
		// does not support explicit suggest triggering, so we fail silently
		// TODO: Revisit once https://github.com/microsoft/language-server-protocol/issues/1117 is addressed
		return nil, nil
	}

	commandPrefix, _ := lsctx.CommandPrefix(ctx)
	handler, ok := cmdHandlers(svc).Get(params.Command, commandPrefix)
	if !ok {
		return nil, fmt.Errorf("%w: command handler not found for %q", jrpc2.MethodNotFound.Err(), params.Command)
	}

	pt, ok := params.WorkDoneToken.(gopls.ProgressToken)
	if ok {
		ctx = lsctx.WithProgressToken(ctx, pt)
	}

	return handler(ctx, cmd.ParseCommandArgs(params.Arguments))
}

func removedHandler(errorMessage string) cmd.Handler {
	return func(context.Context, cmd.CommandArgs) (interface{}, error) {
		return nil, jrpc2.Errorf(jrpc2.MethodNotFound, "REMOVED: %s", errorMessage)
	}
}
