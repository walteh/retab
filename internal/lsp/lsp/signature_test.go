// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lsp

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	gopls "github.com/walteh/retab/gen/gopls/protocol"
)

func TestToSignatureHelp(t *testing.T) {
	testCases := []struct {
		name                  string
		signature             *lang.FunctionSignature
		expectedSignatureHelp *gopls.SignatureHelp
	}{
		{
			"nil",
			nil,
			nil,
		},
		{
			"no parameters",
			&lang.FunctionSignature{
				Name:        "foo() string",
				Description: lang.Markdown("`foo` description"),
			},
			&gopls.SignatureHelp{
				Signatures: []gopls.SignatureInformation{
					{
						Label:         "foo() string",
						Documentation: &gopls.Or_SignatureInformation_documentation{Value: "foo description"},
						Parameters:    []gopls.ParameterInformation{},
					},
				},
			},
		},
		{
			"one parameter",
			&lang.FunctionSignature{
				Name:        "foo(input list of string) map of number",
				Description: lang.Markdown("`foo` description"),
				Parameters: []lang.FunctionParameter{
					{
						Name:        "input",
						Description: lang.Markdown("`input` description"),
					},
				},
				ActiveParameter: 0,
			},
			&gopls.SignatureHelp{
				Signatures: []gopls.SignatureInformation{
					{
						Label:         "foo(input list of string) map of number",
						Documentation: &gopls.Or_SignatureInformation_documentation{Value: "foo description"},
						Parameters: []gopls.ParameterInformation{
							{
								Label:         "input",
								Documentation: "input description",
							},
						},
					},
				},
				ActiveParameter: 0,
				ActiveSignature: 0,
			},
		},
		{
			"multiple parameters",
			&lang.FunctionSignature{
				Name:        "foo(input string, input2 number, input3 string) number",
				Description: lang.Markdown("`foo` description"),
				Parameters: []lang.FunctionParameter{
					{
						Name:        "input",
						Description: lang.Markdown("`input` description"),
					},
					{
						Name:        "input2",
						Description: lang.Markdown("`input2` description"),
					},
					{
						Name:        "input3",
						Description: lang.Markdown("`input3` description"),
					},
				},
				ActiveParameter: 1,
			},
			&gopls.SignatureHelp{
				Signatures: []gopls.SignatureInformation{
					{
						Label:         "foo(input string, input2 number, input3 string) number",
						Documentation: &gopls.Or_SignatureInformation_documentation{Value: "foo description"},
						Parameters: []gopls.ParameterInformation{
							{
								Label:         "input",
								Documentation: "input description",
							},
							{
								Label:         "input2",
								Documentation: "input2 description",
							},
							{
								Label:         "input3",
								Documentation: "input3 description",
							},
						},
					},
				},
				ActiveParameter: 1,
				ActiveSignature: 0,
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%2d-%s", i, tc.name), func(t *testing.T) {
			signature := ToSignatureHelp(tc.signature)

			if diff := cmp.Diff(tc.expectedSignatureHelp, signature); diff != "" {
				t.Fatalf("unexpected signature help: %s", diff)
			}
		})
	}
}
