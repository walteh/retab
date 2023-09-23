// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"
	"io"
	"time"

	"github.com/spf13/afero"
	"github.com/walteh/retab/gen/gopls"
	"github.com/walteh/retab/internal/lsp/document"
	"github.com/walteh/retab/internal/lsp/hcl"
	"github.com/walteh/retab/internal/lsp/lsp"
	"github.com/walteh/retab/pkg/editorconfig"
	"github.com/walteh/retab/pkg/hclwrite"
)

func (svc *service) TextDocumentFormatting(ctx context.Context, params gopls.DocumentFormattingParams) ([]gopls.TextEdit, error) {
	var edits []gopls.TextEdit

	dh := lsp.HandleFromDocumentURI(params.TextDocument.URI)

	filename := string(params.TextDocument.URI)

	text, err := afero.ReadFile(svc.fs, filename)
	if err != nil {
		return edits, err
	}

	edits, err = svc.formatDocument(ctx, text, dh)
	if err != nil {
		return edits, err
	}

	return edits, nil
}

func (svc *service) formatDocument(ctx context.Context, original []byte, dh document.Handle) ([]gopls.TextEdit, error) {

	startTime := time.Now()

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
	svc.logger.Printf("Finished 'formatting' in %s", time.Now().Sub(startTime))

	changes := hcl.Diff(dh, original, formatted)

	return lsp.TextEditsFromDocumentChanges(changes), nil
}
