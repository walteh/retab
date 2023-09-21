// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/walteh/retab/gen/gopls"
	"github.com/walteh/retab/internal/lsp/lsp"
)

func (svc *service) TextDocumentCodeLens(ctx context.Context, params gopls.CodeLensParams) ([]gopls.CodeLens, error) {
	list := make([]gopls.CodeLens, 0)

	dh := lsp.HandleFromDocumentURI(params.TextDocument.URI)
	doc, err := svc.stateStore.DocumentStore.GetDocument(dh)
	if err != nil {
		return list, err
	}

	path := lang.Path{
		Path:       doc.Dir.Path(),
		LanguageID: doc.LanguageID,
	}

	lenses, err := svc.decoder.CodeLensesForFile(ctx, path, doc.Filename)
	if err != nil {
		return nil, err
	}

	for _, lens := range lenses {
		cmd, err := lsp.Command(lens.Command)
		if err != nil {
			svc.logger.Printf("skipping code lens %#v: %s", lens.Command, err)
			continue
		}

		list = append(list, gopls.CodeLens{
			Range:   lsp.HCLRangeToLSP(lens.Range),
			Command: &cmd,
		})
	}

	return list, nil
}
