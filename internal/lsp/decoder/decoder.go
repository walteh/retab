// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/walteh/retab/internal/lsp/codelens"
	"github.com/walteh/retab/internal/lsp/lsp"
	"github.com/walteh/retab/internal/lsp/utm"
	"github.com/walteh/retab/internal/protocol"
)

func varsPathContext(path lang.Path) (*decoder.PathContext, error) {
	// schema, err := tfschema.SchemaForVariables(mod.Meta.Variables, mod.Path)
	// if err != nil {
	// 	return nil, err
	// }

	pathCtx := &decoder.PathContext{
		Schema:           &schema.BodySchema{},
		ReferenceOrigins: make(reference.Origins, 0),
		ReferenceTargets: make(reference.Targets, 0),
		Files:            make(map[string]*hcl.File),
	}

	// for _, origin := range mod.VarsRefOrigins {
	// 	if ast.IsVarsFilename(origin.OriginRange().Filename) {
	// 		pathCtx.ReferenceOrigins = append(pathCtx.ReferenceOrigins, origin)
	// 	}
	// }
	// pathCtx.Files[name] = mod

	return pathCtx, nil
}
func DecoderContext(ctx context.Context) decoder.DecoderContext {
	dCtx := decoder.NewDecoderContext()
	dCtx.UtmSource = utm.UtmSource
	dCtx.UtmMedium = utm.UtmMedium(ctx)
	dCtx.UseUtmContent = true

	cc, err := lsp.ClientCapabilities(ctx)
	if err == nil {
		cmdId, ok := protocol.ExperimentalClientCapabilities(cc.Experimental).ShowReferencesCommandId()
		if ok {
			dCtx.CodeLenses = append(dCtx.CodeLenses, codelens.ReferenceCount(cmdId))
		}
	}

	return dCtx
}
