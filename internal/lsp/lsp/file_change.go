// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lsp

import (
	"github.com/walteh/retab/gen/gopls"
	"github.com/walteh/retab/internal/lsp/document"
)

type contentChange struct {
	text string
	rng  *document.Range
}

func ContentChange(chEvent gopls.TextDocumentContentChangeEvent) document.Change {
	return &contentChange{
		text: chEvent.Text,
		rng:  lspRangeToDocRange(chEvent.Range),
	}
}

func DocumentChanges(events []gopls.TextDocumentContentChangeEvent) document.Changes {
	changes := make(document.Changes, len(events))
	for i, event := range events {
		ch := ContentChange(event)
		changes[i] = ch
	}
	return changes
}

func (fc *contentChange) Text() string {
	return fc.text
}

func (fc *contentChange) Range() *document.Range {
	return fc.rng
}
