// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"
	"path/filepath"

	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/hashicorp/hcl-lang/lang"
	gopls "github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/lsp/lsp"
)

func (svc *service) GoToDefinition(ctx context.Context, params gopls.TextDocumentPositionParams) (interface{}, error) {
	cc, err := lsp.ClientCapabilities(ctx)
	if err != nil {
		return nil, err
	}

	targets, err := svc.goToReferenceTarget(ctx, params)
	if err != nil {
		return nil, err
	}

	return lsp.RefTargetsToDefinitionLocationLinks(targets, cc.TextDocument.Definition), nil
}

func (svc *service) GoToDeclaration(ctx context.Context, params gopls.TextDocumentPositionParams) (interface{}, error) {
	cc, err := lsp.ClientCapabilities(ctx)
	if err != nil {
		return nil, err
	}

	targets, err := svc.goToReferenceTarget(ctx, params)
	if err != nil {
		return nil, err
	}

	return lsp.RefTargetsToDeclarationLocationLinks(targets, cc.TextDocument.Declaration), nil
}

func (svc *service) goToReferenceTarget(ctx context.Context, params gopls.TextDocumentPositionParams) (decoder.ReferenceTargets, error) {
	filename := string(params.TextDocument.URI)

	pos, err := lsp.HCLPositionFromLspPosition(params.Position, svc.fs, filename)
	if err != nil {
		return nil, err
	}

	path := lang.Path{
		Path:       filepath.Dir(filename),
		LanguageID: lsp.Retab.String(),
	}

	return svc.decoder.ReferenceTargetsForOriginAtPos(path, filename, pos)
}
