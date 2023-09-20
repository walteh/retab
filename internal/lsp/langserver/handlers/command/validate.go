// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"context"
	"fmt"

	"github.com/creachadair/jrpc2"
	lsctx "github.com/walteh/retab/internal/lsp/context"
	"github.com/walteh/retab/internal/lsp/document"
	"github.com/walteh/retab/internal/lsp/langserver/cmd"
	"github.com/walteh/retab/internal/lsp/langserver/diagnostics"
	"github.com/walteh/retab/internal/lsp/langserver/errors"
	"github.com/walteh/retab/internal/lsp/langserver/progress"
	"github.com/walteh/retab/internal/lsp/state"
	"github.com/walteh/retab/internal/lsp/terraform/module"
	"github.com/walteh/retab/internal/lsp/uri"
)

func (h *CmdHandler) TerraformValidateHandler(ctx context.Context, args cmd.CommandArgs) (interface{}, error) {
	dirUri, ok := args.GetString("uri")
	if !ok || dirUri == "" {
		return nil, fmt.Errorf("%w: expected module uri argument to be set", jrpc2.InvalidParams.Err())
	}

	if !uri.IsURIValid(dirUri) {
		return nil, fmt.Errorf("URI %q is not valid", dirUri)
	}

	dirHandle := document.DirHandleFromURI(dirUri)

	mod, err := h.StateStore.Modules.ModuleByPath(dirHandle.Path())
	if err != nil {
		if state.IsModuleNotFound(err) {
			err = h.StateStore.Modules.Add(dirHandle.Path())
			if err != nil {
				return nil, err
			}
			mod, err = h.StateStore.Modules.ModuleByPath(dirHandle.Path())
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	tfExec, err := module.TerraformExecutorForModule(ctx, mod.Path)
	if err != nil {
		return nil, errors.EnrichTfExecError(err)
	}

	notifier, err := lsctx.DiagnosticsNotifier(ctx)
	if err != nil {
		return nil, err
	}

	progress.Begin(ctx, "Validating")
	defer func() {
		progress.End(ctx, "Finished")
	}()
	progress.Report(ctx, "Running terraform validate ...")
	jsonDiags, err := tfExec.Validate(ctx)
	if err != nil {
		return nil, err
	}

	diags := diagnostics.NewDiagnostics()
	validateDiags := diagnostics.HCLDiagsFromJSON(jsonDiags)
	diags.EmptyRootDiagnostic()
	diags.Append("terraform validate", validateDiags)
	diags.Append("HCL", mod.ModuleDiagnostics.AutoloadedOnly().AsMap())
	diags.Append("HCL", mod.VarsDiagnostics.AutoloadedOnly().AsMap())

	notifier.PublishHCLDiags(ctx, mod.Path, diags)

	return nil, nil
}