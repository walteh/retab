// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"
	"time"

	lsp "github.com/walteh/retab/gen/gopls"
	"github.com/walteh/retab/internal/lsp/document"
	"github.com/walteh/retab/internal/lsp/hcl"
	"github.com/walteh/retab/internal/lsp/langserver/errors"
	ilsp "github.com/walteh/retab/internal/lsp/lsp"
	"github.com/walteh/retab/internal/lsp/terraform/exec"
	"github.com/walteh/retab/internal/lsp/terraform/module"
)

func (svc *service) TextDocumentFormatting(ctx context.Context, params lsp.DocumentFormattingParams) ([]lsp.TextEdit, error) {
	var edits []lsp.TextEdit

	dh := ilsp.HandleFromDocumentURI(params.TextDocument.URI)

	tfExec, err := module.TerraformExecutorForModule(ctx, dh.Dir.Path())
	if err != nil {
		return edits, errors.EnrichTfExecError(err)
	}

	doc, err := svc.stateStore.DocumentStore.GetDocument(dh)
	if err != nil {
		return edits, err
	}

	edits, err = svc.formatDocument(ctx, tfExec, doc.Text, dh)
	if err != nil {
		return edits, err
	}

	return edits, nil
}

func (svc *service) formatDocument(ctx context.Context, tfExec exec.TerraformExecutor, original []byte, dh document.Handle) ([]lsp.TextEdit, error) {
	var edits []lsp.TextEdit

	svc.logger.Printf("formatting document via %q", tfExec.GetExecPath())

	startTime := time.Now()
	formatted, err := tfExec.Format(ctx, original)
	if err != nil {
		svc.logger.Printf("Failed 'terraform fmt' in %s", time.Now().Sub(startTime))
		return edits, err
	}
	svc.logger.Printf("Finished 'terraform fmt' in %s", time.Now().Sub(startTime))

	changes := hcl.Diff(dh, original, formatted)

	return ilsp.TextEditsFromDocumentChanges(changes), nil
}
