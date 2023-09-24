package cache

import (
	"context"

	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/walteh/retab/internal/lsp/utm"
)

var _ decoder.PathReader = (*Session)(nil)

// PathContext implements decoder.PathReader.
func (me *Session) PathContext(path lang.Path) (*decoder.PathContext, error) {

	return &decoder.PathContext{
		Schema:           &schema.BodySchema{},
		ReferenceOrigins: make(reference.Origins, 0),
		ReferenceTargets: make(reference.Targets, 0),
		Files:            make(map[string]*hcl.File),
	}, nil
}

// Paths implements decoder.PathReader.
func (me *Session) Paths(ctx context.Context) []lang.Path {
	paths := make([]lang.Path, 0, len(me.overlays))
	for _, ov := range me.overlays {
		paths = append(paths, lang.Path{
			Path:       string(ov.URI()),
			LanguageID: "retab",
		})
	}
	return paths
}

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

	return dCtx
}
