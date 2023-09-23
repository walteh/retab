// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lsp

import (
	"github.com/hashicorp/hcl-lang/lang"
	gopls "github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/lsp/mdplain"
)

func ToCompletionList(candidates lang.Candidates, caps gopls.TextDocumentClientCapabilities) gopls.CompletionList {
	list := gopls.CompletionList{
		Items:        make([]gopls.CompletionItem, len(candidates.List)),
		IsIncomplete: !candidates.IsComplete,
	}

	for i, c := range candidates.List {
		list.Items[i] = toCompletionItem(c, caps.Completion)
	}

	return list
}

func toCompletionItem(candidate lang.Candidate, caps gopls.CompletionClientCapabilities) gopls.CompletionItem {
	snippetSupport := caps.CompletionItem.SnippetSupport

	doc := candidate.Description.Value

	// TODO: Revisit when MarkupContent is allowed as Documentation
	// https://github.com/golang/tools/blob/4783bc9b/internal/lsp/protocol/tsprotocol.go#L753
	doc = mdplain.Clean(doc)

	var kind gopls.CompletionItemKind
	switch candidate.Kind {
	case lang.AttributeCandidateKind:
		kind = gopls.PropertyCompletion
	case lang.BlockCandidateKind:
		kind = gopls.ClassCompletion
	case lang.LabelCandidateKind:
		kind = gopls.FieldCompletion
	case lang.BoolCandidateKind:
		kind = gopls.EnumMemberCompletion
	case lang.StringCandidateKind:
		kind = gopls.TextCompletion
	case lang.NumberCandidateKind:
		kind = gopls.ValueCompletion
	case lang.KeywordCandidateKind:
		kind = gopls.KeywordCompletion
	case lang.ListCandidateKind, lang.SetCandidateKind, lang.TupleCandidateKind:
		kind = gopls.EnumCompletion
	case lang.MapCandidateKind, lang.ObjectCandidateKind:
		kind = gopls.StructCompletion
	case lang.TraversalCandidateKind:
		kind = gopls.VariableCompletion
	}

	// TODO: Omit item which uses kind unsupported by the client

	var cmd *gopls.Command
	if candidate.TriggerSuggest && snippetSupport {
		cmd = &gopls.Command{
			Command: "editor.action.triggerSuggest",
			Title:   "Suggest",
		}
	}

	item := gopls.CompletionItem{
		Label:               candidate.Label,
		Kind:                kind,
		InsertTextFormat:    insertTextFormat(snippetSupport),
		Detail:              candidate.Detail,
		Documentation:       &gopls.Or_CompletionItem_documentation{Value: doc},
		TextEdit:            textEdit(candidate.TextEdit, snippetSupport),
		Command:             cmd,
		AdditionalTextEdits: TextEdits(candidate.AdditionalTextEdits, snippetSupport),
		SortText:            candidate.SortText,
	}

	if candidate.ResolveHook != nil {
		item.Data = candidate.ResolveHook
	}

	if caps.CompletionItem.DeprecatedSupport {
		item.Deprecated = candidate.IsDeprecated
	}
	if tagSliceContains(caps.CompletionItem.TagSupport.ValueSet,
		gopls.ComplDeprecated) && candidate.IsDeprecated {
		item.Tags = []gopls.CompletionItemTag{
			gopls.ComplDeprecated,
		}
	}

	return item
}

func tagSliceContains(supported []gopls.CompletionItemTag, tag gopls.CompletionItemTag) bool {
	for _, item := range supported {
		if item == tag {
			return true
		}
	}
	return false
}
