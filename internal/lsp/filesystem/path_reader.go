package filesystem

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/walteh/retab/internal/lsp/document"
	"github.com/walteh/retab/internal/lsp/lsp"
	"github.com/walteh/retab/pkg/hclread"
)

var _ decoder.PathReader = &Filesystem{}

func (mr *Filesystem) Paths(ctx context.Context) []lang.Path {
	paths := make([]lang.Path, 0)

	docList, err := mr.docStore.ListDocumentsInDir(document.DirHandleFromPath("."))
	if err != nil {
		return paths
	}

	filtered := make([]*document.Document, 0)
	for _, doc := range docList {
		if doc.LanguageID == lsp.Retab.String() {
			filtered = append(filtered, doc)
		}
	}

	for _, doc := range filtered {
		paths = append(paths, lang.Path{
			Path:       doc.Dir.Path(),
			LanguageID: doc.LanguageID,
		})
	}

	return paths
}

func (fs *Filesystem) PathContext(path lang.Path) (*decoder.PathContext, error) {

	dirHandle := document.DirHandleFromPath(path.Path)
	docList, err := fs.docStore.ListDocumentsInDir(dirHandle)
	if err != nil {
		return nil, fmt.Errorf("doc FS: %w", err)
	}

	filtered := make([]*document.Document, 0)
	for _, doc := range docList {
		if doc.LanguageID == path.LanguageID {
			filtered = append(filtered, doc)
		}
	}

	pathCtx := &decoder.PathContext{
		Schema:           &schema.BodySchema{},
		ReferenceOrigins: make(reference.Origins, 0),
		ReferenceTargets: make(reference.Targets, 0),
		Files:            make(map[string]*hcl.File),
	}

	for _, doc := range filtered {
		fle, err := doc.AsFile()
		if err != nil {
			return nil, fmt.Errorf("doc FS: %w", err)
		}

		fil, _, _, err := hclread.NewEvaluation(context.Background(), fle)
		if err != nil {
			return nil, fmt.Errorf("doc FS: %w", err)
		}
		pathCtx.Files[doc.FullPath()] = fil
	}

	// fmt.Errorf("unknown language ID: %q", path.LanguageID)
	return pathCtx, nil
}
