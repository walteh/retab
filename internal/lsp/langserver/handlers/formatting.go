// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"
	"io"
	"time"

	lsp "github.com/walteh/retab/gen/gopls"
	"github.com/walteh/retab/internal/lsp/document"
	"github.com/walteh/retab/internal/lsp/hcl"
	ilsp "github.com/walteh/retab/internal/lsp/lsp"
	"github.com/walteh/retab/pkg/editorconfig"
	"github.com/walteh/retab/pkg/hclwrite"
)

func (svc *service) TextDocumentFormatting(ctx context.Context, params lsp.DocumentFormattingParams) ([]lsp.TextEdit, error) {
	var edits []lsp.TextEdit

	dh := ilsp.HandleFromDocumentURI(params.TextDocument.URI)

	doc, err := svc.stateStore.DocumentStore.GetDocument(dh)
	if err != nil {
		return edits, err
	}

	edits, err = svc.formatDocument(ctx, doc.Text, dh)
	if err != nil {
		return edits, err
	}

	return edits, nil
}

func (svc *service) formatDocument(ctx context.Context, original []byte, dh document.Handle) ([]lsp.TextEdit, error) {

	startTime := time.Now()
	// formatted, err := tfExec.Format(ctx, original)
	// if err != nil {
	// 	svc.logger.Printf("Failed 'terraform fmt' in %s", time.Now().Sub(startTime))
	// 	return edits, err
	// }

	cfg, err := editorconfig.NewEditorConfigConfigurationProvider(ctx, dh.Dir.URI)
	if err != nil {
		return nil, err
	}

	r, err := hclwrite.FormatBytes(cfg, original)
	if err != nil {
		return nil, err
	}
	formatted, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	svc.logger.Printf("Finished 'terraform fmt' in %s", time.Now().Sub(startTime))

	changes := hcl.Diff(dh, original, formatted)

	return ilsp.TextEditsFromDocumentChanges(changes), nil
}
