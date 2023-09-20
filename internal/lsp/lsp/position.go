// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lsp

import (
	"github.com/hashicorp/hcl/v2"
	lsp "github.com/walteh/retab/gen/gopls"
	"github.com/walteh/retab/internal/lsp/document"
)

func HCLPositionFromLspPosition(pos lsp.Position, doc *document.Document) (hcl.Pos, error) {
	byteOffset, err := document.ByteOffsetForPos(doc.Lines, lspPosToDocumentPos(pos))
	if err != nil {
		return hcl.Pos{}, err
	}

	return hcl.Pos{
		Line:   int(pos.Line) + 1,
		Column: int(pos.Character) + 1,
		Byte:   byteOffset,
	}, nil
}

func lspPosToDocumentPos(pos lsp.Position) document.Pos {
	return document.Pos{
		Line:   int(pos.Line),
		Column: int(pos.Character),
	}
}
