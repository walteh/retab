package lsp

import (
	"context"

	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/afero"
	"github.com/walteh/retab/pkg/hclread"
)

type pathReader struct {
	basefs afero.Fs
}

func NewPathReader(basefs afero.Fs, path string) decoder.PathReader {
	return &pathReader{
		basefs: afero.NewBasePathFs(afero.NewReadOnlyFs(basefs), path),
	}
}

var _ decoder.PathReader = (*pathReader)(nil)

// PathContext implements decoder.PathReader.
func (me *pathReader) PathContext(path lang.Path) (*decoder.PathContext, error) {
	pctx := &decoder.PathContext{
		Files: map[string]*hcl.File{},
	}

	fls, err := afero.ReadDir(me.basefs, path.Path)
	if err != nil {
		return nil, err
	}

	for _, fl := range fls {
		fle, err := me.basefs.Open(fl.Name())
		if err != nil {
			return nil, err
		}
		f, _, _, err := hclread.NewEvaluation(context.Background(), fle)
		if err != nil {
			return nil, err
		}

		pctx.Files[path.Path] = f
	}

	return pctx, nil

}

// Paths implements decoder.PathReader.
func (me *pathReader) Paths(ctx context.Context) []lang.Path {

	fls, err := afero.ReadDir(me.basefs, ".")
	if err != nil {
		return nil
	}

	paths := make([]lang.Path, 0, len(fls))

	for _, fl := range fls {
		paths = append(paths, lang.Path{
			Path:       fl.Name(),
			LanguageID: "retab",
		})
	}

	return paths

}
