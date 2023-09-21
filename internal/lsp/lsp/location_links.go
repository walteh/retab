// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lsp

import (
	"path/filepath"

	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/walteh/retab/gen/gopls"
	"github.com/walteh/retab/internal/lsp/uri"
)

func RefTargetsToDefinitionLocationLinks(targets decoder.ReferenceTargets, defCaps *gopls.DefinitionClientCapabilities) interface{} {
	if defCaps == nil {
		return RefTargetsToLocationLinks(targets, false)
	}
	return RefTargetsToLocationLinks(targets, defCaps.LinkSupport)
}

func RefTargetsToDeclarationLocationLinks(targets decoder.ReferenceTargets, declCaps *gopls.DeclarationClientCapabilities) interface{} {
	if declCaps == nil {
		return RefTargetsToLocationLinks(targets, false)
	}
	return RefTargetsToLocationLinks(targets, declCaps.LinkSupport)
}

func RefTargetsToLocationLinks(targets decoder.ReferenceTargets, linkSupport bool) interface{} {
	if linkSupport {
		links := make([]gopls.LocationLink, 0)
		for _, target := range targets {
			links = append(links, refTargetToLocationLink(target))
		}
		return links
	}

	locations := make([]gopls.Location, 0)
	for _, target := range targets {
		locations = append(locations, refTargetToLocation(target))
	}
	return locations
}

func refTargetToLocationLink(target *decoder.ReferenceTarget) gopls.LocationLink {
	targetUri := uri.FromPath(filepath.Join(target.Path.Path, target.Range.Filename))
	originRange := HCLRangeToLSP(target.OriginRange)

	locLink := gopls.LocationLink{
		OriginSelectionRange: &originRange,
		TargetURI:            gopls.DocumentURI(targetUri),
		TargetRange:          HCLRangeToLSP(target.Range),
		TargetSelectionRange: HCLRangeToLSP(target.Range),
	}

	if target.DefRangePtr != nil {
		locLink.TargetSelectionRange = HCLRangeToLSP(*target.DefRangePtr)
	}

	return locLink
}

func refTargetToLocation(target *decoder.ReferenceTarget) gopls.Location {
	targetUri := uri.FromPath(filepath.Join(target.Path.Path, target.Range.Filename))

	return gopls.Location{
		URI:   gopls.DocumentURI(targetUri),
		Range: HCLRangeToLSP(target.Range),
	}
}
