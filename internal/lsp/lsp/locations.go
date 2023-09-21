// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lsp

import (
	"path/filepath"

	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/walteh/retab/gen/gopls"
	"github.com/walteh/retab/internal/lsp/uri"
)

func RefOriginsToLocations(origins decoder.ReferenceOrigins) []gopls.Location {
	locations := make([]gopls.Location, len(origins))

	for i, origin := range origins {
		originUri := uri.FromPath(filepath.Join(origin.Path.Path, origin.Range.Filename))
		locations[i] = gopls.Location{
			URI:   gopls.DocumentURI(originUri),
			Range: HCLRangeToLSP(origin.Range),
		}
	}

	return locations
}
