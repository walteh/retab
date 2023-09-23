// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lsp

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/afero"
	gopls "github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/lsp/document"
	"github.com/walteh/retab/internal/lsp/source"
)

func HCLPositionFromLspPosition(pos gopls.Position, fle afero.Fs, name string) (hcl.Pos, error) {

	all, err := afero.ReadFile(fle, name)
	if err != nil {
		return hcl.Pos{}, err
	}

	lines := source.MakeSourceLines(name, all)

	byteOffset, err := document.ByteOffsetForPos(lines, lspPosToDocumentPos(pos))
	if err != nil {
		return hcl.Pos{}, err
	}

	return hcl.Pos{
		Line:   int(pos.Line) + 1,
		Column: int(pos.Character) + 1,
		Byte:   byteOffset,
	}, nil
}

func lspPosToDocumentPos(pos gopls.Position) document.Pos {
	return document.Pos{
		Line:   int(pos.Line),
		Column: int(pos.Character),
	}
}
