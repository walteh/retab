package filesystem

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/walteh/retab/internal/lsp/lsp"
	"github.com/walteh/retab/pkg/hclread"
)

var _ decoder.PathReader = &Filesystem{}

func (me *Filesystem) Paths(ctx context.Context) []lang.Path {
	paths := make([]lang.Path, 0)

	docList, err := afero.ReadDir(me, ".")
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("doc FS")
		return paths
	}

	// filtered := make([]*document.Document, 0)
	// for _, doc := range docList {
	// 	if doc.LanguageID == lsp.Retab.String() {
	// 		filtered = append(filtered, doc)
	// 	}
	// }

	for _, doc := range docList {
		paths = append(paths, lang.Path{
			Path:       doc.Name(),
			LanguageID: lsp.Retab.String(),
		})
	}

	return paths
}

func (me *Filesystem) PathContext(path lang.Path) (*decoder.PathContext, error) {

	docList, err := afero.ReadDir(me, path.Path)
	if err != nil {
		return nil, fmt.Errorf("doc FS: %w", err)
	}

	// filtered := make([]*document.Document, 0)
	// for _, doc := range docList {
	// 	if doc.Name() == path.LanguageID {
	// 		filtered = append(filtered, doc)
	// 	}
	// }

	pathCtx := &decoder.PathContext{
		Schema:           &schema.BodySchema{},
		ReferenceOrigins: make(reference.Origins, 0),
		ReferenceTargets: make(reference.Targets, 0),
		Files:            make(map[string]*hcl.File),
	}

	for _, doc := range docList {
		// fle, err := doc.AsFile()
		// if err != nil {
		// 	return nil, fmt.Errorf("doc FS: %w", err)
		// }

		fle, err := me.Open(doc.Name())
		if err != nil {
			return nil, fmt.Errorf("doc FS: %w", err)
		}

		fil, _, _, err := hclread.NewEvaluation(context.Background(), fle)
		if err != nil {
			return nil, fmt.Errorf("doc FS: %w", err)
		}
		pathCtx.Files[fle.Name()] = fil
	}

	// fmt.Errorf("unknown language ID: %q", path.LanguageID)
	return pathCtx, nil
}
