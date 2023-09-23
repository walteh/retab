// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"
	"fmt"

	"github.com/spf13/afero"
	gopls "github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/lsp/lsp"
)

func (svc *service) TextDocumentCodeAction(ctx context.Context, params gopls.CodeActionParams) []gopls.CodeAction {
	ca, err := svc.textDocumentCodeAction(ctx, params)
	if err != nil {
		svc.logger.Printf("code action failed: %s", err)
	}

	return ca
}

func (svc *service) textDocumentCodeAction(ctx context.Context, params gopls.CodeActionParams) ([]gopls.CodeAction, error) {
	var ca []gopls.CodeAction

	// For action definitions, refer to https://code.visualstudio.com/api/references/vscode-api#CodeActionKind
	// We only support format type code actions at the moment, and do not want to format without the client asking for
	// them, so exit early here if nothing is requested.
	if len(params.Context.Only) == 0 {
		svc.logger.Printf("No code action requested, exiting")
		return ca, nil
	}

	for _, o := range params.Context.Only {
		svc.logger.Printf("Code actions requested: %q", o)
	}

	wantedCodeActions := lsp.SupportedCodeActions.Only(params.Context.Only)
	if len(wantedCodeActions) == 0 {
		return nil, fmt.Errorf("could not find a supported code action to execute for %s, wanted %v",
			params.TextDocument.URI, params.Context.Only)
	}

	svc.logger.Printf("Code actions supported: %v", wantedCodeActions)

	filename := string(params.TextDocument.URI)

	text, err := afero.ReadFile(svc.fs, filename)
	if err != nil {
		return ca, err
	}

	for action := range wantedCodeActions {
		switch action {
		case lsp.SourceFormatAllTerraform:

			edits, err := svc.formatDocument(ctx, text, filename)
			if err != nil {
				return ca, err
			}

			ca = append(ca, gopls.CodeAction{
				Title: "Format Document",
				Kind:  action,
				Edit: &gopls.WorkspaceEdit{
					Changes: map[gopls.DocumentURI][]gopls.TextEdit{
						gopls.DocumentURI(filename): edits,
					},
				},
			})
		}
	}

	return ca, nil
}
