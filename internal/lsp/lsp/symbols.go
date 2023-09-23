// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lsp

import (
	"path/filepath"

	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/hashicorp/hcl-lang/lang"
	gopls "github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/lsp/uri"
	"github.com/zclconf/go-cty/cty"
)

func WorkspaceSymbols(sbs []decoder.Symbol, caps *gopls.WorkspaceSymbolClientCapabilities) []gopls.SymbolInformation {
	symbols := make([]gopls.SymbolInformation, len(sbs))
	for i, s := range sbs {
		kind, ok := symbolKind(s, caps.SymbolKind.ValueSet)
		if !ok {
			// skip symbol not supported by client
			continue
		}

		path := filepath.Join(s.Path().Path, s.Range().Filename)
		symbols[i] = gopls.SymbolInformation{
			Name: s.Name(),
			Kind: kind,
			Location: gopls.Location{
				Range: HCLRangeToLSP(s.Range()),
				URI:   gopls.DocumentURI(uri.FromPath(path)),
			},
		}
	}
	return symbols
}

func DocumentSymbols(sbs []decoder.Symbol, caps gopls.DocumentSymbolClientCapabilities) []gopls.DocumentSymbol {
	symbols := make([]gopls.DocumentSymbol, 0)

	for _, s := range sbs {
		symbol, ok := documentSymbol(s, caps)
		if !ok {
			// skip symbol not supported by client
			continue
		}
		symbols = append(symbols, symbol)
	}
	return symbols
}

func documentSymbol(symbol decoder.Symbol, caps gopls.DocumentSymbolClientCapabilities) (gopls.DocumentSymbol, bool) {
	kind, ok := symbolKind(symbol, caps.SymbolKind.ValueSet)
	if !ok {
		return gopls.DocumentSymbol{}, false
	}

	ds := gopls.DocumentSymbol{
		Name:           symbol.Name(),
		Kind:           kind,
		Range:          HCLRangeToLSP(symbol.Range()),
		SelectionRange: HCLRangeToLSP(symbol.Range()),
	}
	if caps.HierarchicalDocumentSymbolSupport {
		ds.Children = DocumentSymbols(symbol.NestedSymbols(), caps)
	}
	return ds, true
}

func symbolKind(symbol decoder.Symbol, supported []gopls.SymbolKind) (gopls.SymbolKind, bool) {
	switch s := symbol.(type) {
	case *decoder.BlockSymbol:
		kind, ok := supportedSymbolKind(supported, gopls.Class)
		if ok {
			return kind, true
		}
	case *decoder.AttributeSymbol:
		kind, ok := exprSymbolKind(s.ExprKind, supported)
		if ok {
			return kind, true
		}
	case *decoder.ExprSymbol:
		kind, ok := exprSymbolKind(s.ExprKind, supported)
		if ok {
			return kind, true
		}
	}

	return gopls.SymbolKind(0), false
}

func exprSymbolKind(symbolKind lang.SymbolExprKind, supported []gopls.SymbolKind) (gopls.SymbolKind, bool) {
	switch k := symbolKind.(type) {
	case lang.LiteralTypeKind:
		switch k.Type {
		case cty.Bool:
			return supportedSymbolKind(supported, gopls.Boolean)
		case cty.String:
			return supportedSymbolKind(supported, gopls.String)
		case cty.Number:
			return supportedSymbolKind(supported, gopls.Number)
		}
	case lang.TraversalExprKind:
		return supportedSymbolKind(supported, gopls.Constant)
	case lang.TupleConsExprKind:
		return supportedSymbolKind(supported, gopls.Array)
	case lang.ObjectConsExprKind:
		return supportedSymbolKind(supported, gopls.Struct)
	}

	return supportedSymbolKind(supported, gopls.Variable)
}

func supportedSymbolKind(supported []gopls.SymbolKind, kind gopls.SymbolKind) (gopls.SymbolKind, bool) {
	for _, s := range supported {
		if s == kind {
			return s, true
		}
	}
	return gopls.SymbolKind(0), false
}
