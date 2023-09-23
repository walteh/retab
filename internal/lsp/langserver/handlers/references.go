// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"
	"path/filepath"

	"github.com/hashicorp/hcl-lang/lang"
	gopls "github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/lsp/lsp"
)

func (svc *service) References(ctx context.Context, params gopls.ReferenceParams) ([]gopls.Location, error) {
	list := make([]gopls.Location, 0)

	filename := string(params.TextDocument.URI)

	pos, err := lsp.HCLPositionFromLspPosition(params.TextDocumentPositionParams.Position, svc.fs, filename)
	if err != nil {
		return list, err
	}

	path := lang.Path{
		Path:       filepath.Dir(filename),
		LanguageID: lsp.Retab.String(),
	}

	origins := svc.decoder.ReferenceOriginsTargetingPos(path, filename, pos)

	return lsp.RefOriginsToLocations(origins), nil
}
