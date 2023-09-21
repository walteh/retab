// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/walteh/retab/gen/gopls"
	"github.com/walteh/retab/internal/lsp/lsp"
)

func (svc *service) References(ctx context.Context, params gopls.ReferenceParams) ([]gopls.Location, error) {
	list := make([]gopls.Location, 0)

	dh := lsp.HandleFromDocumentURI(params.TextDocument.URI)
	doc, err := svc.stateStore.DocumentStore.GetDocument(dh)
	if err != nil {
		return list, err
	}

	pos, err := lsp.HCLPositionFromLspPosition(params.TextDocumentPositionParams.Position, doc)
	if err != nil {
		return list, err
	}

	path := lang.Path{
		Path:       doc.Dir.Path(),
		LanguageID: doc.LanguageID,
	}

	origins := svc.decoder.ReferenceOriginsTargetingPos(path, doc.Filename, pos)

	return lsp.RefOriginsToLocations(origins), nil
}
