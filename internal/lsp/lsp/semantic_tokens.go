// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lsp

import (
	"github.com/hashicorp/hcl-lang/lang"
	gopls "github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/lsp/lsp/semtok"
	"github.com/walteh/retab/pkg/tokmod"
)

// Registering types which are actually in use
var (
	serverTokenTypes = semtok.TokenTypes{
		semtok.TokenTypeEnumMember,
		semtok.TokenTypeFunction,
		semtok.TokenTypeKeyword,
		semtok.TokenTypeNumber,
		semtok.TokenTypeParameter,
		semtok.TokenTypeProperty,
		semtok.TokenTypeString,
		semtok.TokenTypeType,
		semtok.TokenTypeVariable,
	}
	serverTokenModifiers = semtok.TokenModifiers{
		semtok.TokenModifierDefaultLibrary,
	}
)

func init() {
	for _, tokType := range lang.SupportedSemanticTokenTypes {
		serverTokenTypes = append(serverTokenTypes, semtok.TokenType(tokType))
	}
	serverTokenModifiers = append(serverTokenModifiers, semtok.TokenModifier(lang.TokenModifierDependent))
	for _, tokModifier := range tokmod.SupportedModifiers {
		serverTokenModifiers = append(serverTokenModifiers, semtok.TokenModifier(tokModifier))
	}
}

func TokenTypesLegend(clientSupported []string) semtok.TokenTypes {
	legend := make(semtok.TokenTypes, 0)

	// Filter only supported token types
	for _, tokenType := range serverTokenTypes {
		if sliceContains(clientSupported, string(tokenType)) {
			legend = append(legend, semtok.TokenType(tokenType))
		}
	}

	return legend
}

func TokenModifiersLegend(clientSupported []string) semtok.TokenModifiers {
	legend := make(semtok.TokenModifiers, 0)

	// Filter only supported token modifiers
	for _, modifier := range serverTokenModifiers {
		if sliceContains(clientSupported, string(modifier)) {
			legend = append(legend, semtok.TokenModifier(modifier))
		}
	}

	return legend
}

type SemanticTokensClientCapabilities struct {
	gopls.SemanticTokensClientCapabilities
}

func (c SemanticTokensClientCapabilities) FullRequest() bool {
	switch full := c.Requests.Full.Value.(type) {
	case bool:
		return full
	case map[string]interface{}:
		return true
	}
	return false
}
