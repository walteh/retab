// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	gopls "github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/lsp/lsp"
)

func (svc *service) TextDocumentCodeLens(ctx context.Context, params gopls.CodeLensParams) ([]gopls.CodeLens, error) {
	list := make([]gopls.CodeLens, 0)

	filename := string(params.TextDocument.URI)

	path := lang.Path{
		Path:       filename,
		LanguageID: lsp.Retab.String(),
	}

	lenses, err := svc.decoder.CodeLensesForFile(ctx, path, filename)
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
