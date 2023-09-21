// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lsp

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/walteh/retab/gen/gopls"
	"github.com/walteh/retab/internal/lsp/document"
)

func TextEditsFromDocumentChanges(changes document.Changes) []gopls.TextEdit {
	edits := make([]gopls.TextEdit, len(changes))

	for i, change := range changes {
		edits[i] = gopls.TextEdit{
			Range:   documentRangeToLSP(change.Range()),
			NewText: change.Text(),
		}
	}

	return edits
}

func TextEdits(tes []lang.TextEdit, snippetSupport bool) []gopls.TextEdit {
	edits := make([]gopls.TextEdit, len(tes))

	for i, te := range tes {
		edits[i] = *textEdit(te, snippetSupport)
	}

	return edits
}

func textEdit(te lang.TextEdit, snippetSupport bool) *gopls.TextEdit {
	if snippetSupport {
		return &gopls.TextEdit{
			NewText: te.Snippet,
			Range:   HCLRangeToLSP(te.Range),
		}
	}

	return &gopls.TextEdit{
		NewText: te.NewText,
		Range:   HCLRangeToLSP(te.Range),
	}
}

func insertTextFormat(snippetSupport bool) *gopls.InsertTextFormat {
	dat := gopls.PlainTextTextFormat
	if snippetSupport {
		dat = gopls.SnippetTextFormat
	}

	return &dat
}
