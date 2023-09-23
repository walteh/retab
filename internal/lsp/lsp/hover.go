// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lsp

import (
	"github.com/hashicorp/hcl-lang/lang"
	gopls "github.com/walteh/retab/gen/gopls/protocol"
)

func HoverData(data *lang.HoverData, cc gopls.TextDocumentClientCapabilities) *gopls.Hover {
	if data == nil {
		return nil
	}
	mdSupported := len(cc.Hover.ContentFormat) > 0 && cc.Hover.ContentFormat[0] == "markdown"

	// In theory we should be sending lsp.MarkedString (for old clients)
	// when len(cc.Hover.ContentFormat) == 0, but that's not possible
	// without changing lsp.Hover.Content field type to interface{}
	//
	// We choose to follow gopls' approach (i.e. cut off old clients).

	return &gopls.Hover{
		Contents: markupContent(data.Content, mdSupported),
		Range:    HCLRangeToLSP(data.Range),
	}
}
