// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lsp

import (
	"github.com/hashicorp/hcl/v2"
	gopls "github.com/walteh/retab/gen/gopls/protocol"
)

func HCLSeverityToLSP(severity hcl.DiagnosticSeverity) gopls.DiagnosticSeverity {
	var sev gopls.DiagnosticSeverity
	switch severity {
	case hcl.DiagError:
		sev = gopls.SeverityError
	case hcl.DiagWarning:
		sev = gopls.SeverityWarning
	case hcl.DiagInvalid:
		panic("invalid diagnostic")
	}
	return sev
}

func HCLDiagsToLSP(hclDiags hcl.Diagnostics, source string) []gopls.Diagnostic {
	diags := []gopls.Diagnostic{}

	for _, hclDiag := range hclDiags {
		msg := hclDiag.Summary
		if hclDiag.Detail != "" {
			msg += ": " + hclDiag.Detail
		}
		var rnge gopls.Range
		if hclDiag.Subject != nil {
			rnge = HCLRangeToLSP(*hclDiag.Subject)
		}
		diags = append(diags, gopls.Diagnostic{
			Range:    rnge,
			Severity: HCLSeverityToLSP(hclDiag.Severity),
			Source:   source,
			Message:  msg,
		})

	}
	return diags
}
