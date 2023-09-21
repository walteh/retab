// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lsp

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/walteh/retab/gen/gopls"
	"github.com/walteh/retab/internal/lsp/document"
)

func documentRangeToLSP(docRng *document.Range) gopls.Range {
	if docRng == nil {
		return gopls.Range{}
	}

	return gopls.Range{
		Start: gopls.Position{
			Character: uint32(docRng.Start.Column),
			Line:      uint32(docRng.Start.Line),
		},
		End: gopls.Position{
			Character: uint32(docRng.End.Column),
			Line:      uint32(docRng.End.Line),
		},
	}
}

func lspRangeToDocRange(rng *gopls.Range) *document.Range {
	if rng == nil {
		return nil
	}

	return &document.Range{
		Start: document.Pos{
			Line:   int(rng.Start.Line),
			Column: int(rng.Start.Character),
		},
		End: document.Pos{
			Line:   int(rng.End.Line),
			Column: int(rng.End.Character),
		},
	}
}

func HCLRangeToLSP(rng hcl.Range) gopls.Range {
	return gopls.Range{
		Start: HCLPosToLSP(rng.Start),
		End:   HCLPosToLSP(rng.End),
	}
}

func HCLPosToLSP(pos hcl.Pos) gopls.Position {
	return gopls.Position{
		Line:      uint32(pos.Line - 1),
		Character: uint32(pos.Column - 1),
	}
}
