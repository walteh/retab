// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lsp

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/walteh/retab/gen/gopls"
)

func Links(links []lang.Link, caps *gopls.DocumentLinkClientCapabilities) []gopls.DocumentLink {
	docLinks := make([]gopls.DocumentLink, len(links))

	for i, link := range links {
		tooltip := ""
		if caps != nil && caps.TooltipSupport {
			tooltip = link.Tooltip
		}
		docLinks[i] = gopls.DocumentLink{
			Range:   HCLRangeToLSP(link.Range),
			Target:  &link.URI,
			Tooltip: tooltip,
		}
	}

	return docLinks
}
