// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	"github.com/hashicorp/hcl-lang/decoder"
	gopls "github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/lsp/lsp"
	"github.com/walteh/retab/internal/lsp/mdplain"
	"github.com/walteh/retab/internal/lsp/protocol"
)

func (svc *service) CompletionItemResolve(ctx context.Context, params protocol.CompletionItemWithResolveHook) (protocol.CompletionItemWithResolveHook, error) {
	cc, err := lsp.ClientCapabilities(ctx)
	if err != nil {
		return params, err
	}

	if params.ResolveHook == nil {
		return params, nil
	}

	unresolvedCandidate := decoder.UnresolvedCandidate{
		ResolveHook: params.ResolveHook,
	}

	resolvedCandidate, err := svc.decoder.ResolveCandidate(ctx, unresolvedCandidate)
	if err != nil || resolvedCandidate == nil {
		return params, err
	}

	if resolvedCandidate.Description.Value != "" {
		doc := resolvedCandidate.Description.Value

		// TODO: Revisit when MarkupContent is allowed as Documentation
		// https://github.com/golang/tools/blob/4783bc9b/internal/lsp/protocol/tsprotocol.go#L753
		doc = mdplain.Clean(doc)
		params.Documentation = &gopls.Or_CompletionItem_documentation{
			Value: doc,
		}
	}
	if resolvedCandidate.Detail != "" {
		params.Detail = resolvedCandidate.Detail
	}
	if len(resolvedCandidate.AdditionalTextEdits) > 0 {
		snippetSupport := cc.TextDocument.Completion.CompletionItem.SnippetSupport
		params.AdditionalTextEdits = lsp.TextEdits(resolvedCandidate.AdditionalTextEdits, snippetSupport)
	}

	return params, nil
}
