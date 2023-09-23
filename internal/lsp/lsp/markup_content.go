// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lsp

import (
	"github.com/hashicorp/hcl-lang/lang"
	gopls "github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/lsp/mdplain"
)

func markupContent(content lang.MarkupContent, mdSupported bool) gopls.MarkupContent {
	value := content.Value

	kind := gopls.PlainText
	if content.Kind == lang.MarkdownKind {
		if mdSupported {
			kind = gopls.Markdown
		} else {
			value = mdplain.Clean(value)
		}
	}

	return gopls.MarkupContent{
		Kind:  kind,
		Value: value,
	}
}
