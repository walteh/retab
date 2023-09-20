// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"
	"fmt"

	"github.com/walteh/retab/internal/lsp/langserver/diagnostics"
	"github.com/walteh/retab/internal/lsp/langserver/notifier"
	"github.com/walteh/retab/internal/lsp/langserver/session"
	"github.com/walteh/retab/internal/lsp/state"
	"github.com/walteh/retab/internal/lsp/telemetry"
)

func sendModuleTelemetry(store *state.StateStore, telemetrySender telemetry.Sender) notifier.Hook {
	return func(ctx context.Context) error {

		return nil
	}
}

func updateDiagnostics(dNotifier *diagnostics.Notifier) notifier.Hook {
	return func(ctx context.Context) error {

		diags := diagnostics.NewDiagnostics()
		diags.EmptyRootDiagnostic()

		// defer dNotifier.PublishHCLDiags(ctx, mod.Path, diags)

		// if mod != nil {
		// 	// diags.Append("HCL", mod.ModuleDiagnostics.AutoloadedOnly().AsMap())
		// 	// diags.Append("HCL", mod.VarsDiagnostics.AutoloadedOnly().AsMap())
		// }

		return nil
	}
}

func callRefreshClientCommand(clientRequester session.ClientCaller, commandId string) notifier.Hook {
	return func(ctx context.Context) error {

		_, err := clientRequester.Callback(ctx, commandId, nil)
		if err != nil {
			return fmt.Errorf("Error calling %s for %s: %s", commandId, ctx.Value("module"), err)
		}

		return nil
	}
}

func refreshCodeLens(clientRequester session.ClientCaller) notifier.Hook {
	return func(ctx context.Context) error {
		// TODO: avoid triggering for new targets outside of open module
		_, err := clientRequester.Callback(ctx, "workspace/codeLens/refresh", nil)
		if err != nil {
			return err
		}
		return nil
	}
}

func refreshSemanticTokens(clientRequester session.ClientCaller) notifier.Hook {
	return func(ctx context.Context) error {

		return nil
	}
}
