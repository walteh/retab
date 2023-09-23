// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lsp

import (
	"github.com/hashicorp/hcl-lang/lang"
	gopls "github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/lsp/mdplain"
)

func ToSignatureHelp(signature *lang.FunctionSignature) *gopls.SignatureHelp {
	if signature == nil {
		return nil
	}

	parameters := make([]gopls.ParameterInformation, 0, len(signature.Parameters))
	for _, p := range signature.Parameters {
		parameters = append(parameters, gopls.ParameterInformation{
			Label: p.Name,
			// TODO: Support markdown per https://github.com/hashicorp/terraform-ls/issues/1212
			Documentation: mdplain.Clean(p.Description.Value),
		})
	}

	return &gopls.SignatureHelp{
		Signatures: []gopls.SignatureInformation{
			{
				Label: signature.Name,
				// TODO: Support markdown per https://github.com/hashicorp/terraform-ls/issues/1212
				Documentation: &gopls.Or_SignatureInformation_documentation{
					Value: mdplain.Clean(signature.Description.Value),
				},
				Parameters: parameters,
			},
		},
		ActiveParameter: signature.ActiveParameter,
		ActiveSignature: 0,
	}
}
