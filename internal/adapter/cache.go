package adapter

import (
	"bytes"
	"context"
	"path/filepath"

	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/walteh/retab/gen/gopls/cache"
	"github.com/walteh/retab/gen/gopls/span"
	"github.com/walteh/retab/internal/source"
	"github.com/walteh/retab/pkg/hclread"
)

var _ decoder.PathReader = &CacheSnapshotDecoder{}
var _ decoder.PathReader = &CacheSessionDecoder{}

type CacheSessionDecoder struct {
	session *cache.Session
}

func NewCacheSessionDecoder(session *cache.Session) *CacheSessionDecoder {
	return &CacheSessionDecoder{session: session}
}

func NewCacheSnapshotDecoder(snapshot source.Snapshot) *CacheSnapshotDecoder {
	return &CacheSnapshotDecoder{snapshot: snapshot}
}

// PathContext implements decoder.PathReader.
func (me *CacheSessionDecoder) PathContext(path lang.Path) (*decoder.PathContext, error) {
	pctx := &decoder.PathContext{
		Schema:           &schema.BodySchema{},
		ReferenceOrigins: make(reference.Origins, 0),
		ReferenceTargets: make(reference.Targets, 0),
		Files:            make(map[string]*hcl.File),
	}

	fle, err := me.session.ReadFile(context.Background(), span.URIFromPath(path.Path))
	if err != nil {
		return nil, err
	}

	cnt, err := fle.Content()
	if err != nil {
		return nil, err
	}

	fi, _, _, err := hclread.NewEvaluationReadCloser(context.Background(), bytes.NewReader(cnt), filepath.Base(path.Path))
	if err != nil {
		return nil, err
	}

	pctx.Files[path.Path] = fi

	return pctx, nil
}

// Paths implements decoder.PathReader.
func (me *CacheSessionDecoder) Paths(ctx context.Context) []lang.Path {
	views := me.session.Views()
	paths := make([]lang.Path, 0, len(views))
	for _, uri := range views {
		snap, _, err := uri.Snapshot()
		if err != nil {
			continue
		}

		paths = append(paths, (&CacheSnapshotDecoder{snapshot: snap}).Paths(ctx)...)
	}

	return paths
}

type CacheSnapshotDecoder struct {
	snapshot source.Snapshot
}

// PathContext implements decoder.PathReader.
func (me *CacheSnapshotDecoder) PathContext(path lang.Path) (*decoder.PathContext, error) {
	ctx := context.Background()

	pctx := &decoder.PathContext{
		Schema:           &schema.BodySchema{},
		ReferenceOrigins: make(reference.Origins, 0),
		ReferenceTargets: make(reference.Targets, 0),
		Files:            make(map[string]*hcl.File),
	}

	fle, err := me.snapshot.ReadFile(context.Background(), span.URIFromPath(path.Path))
	if err != nil {
		return nil, err
	}

	cnt, err := fle.Content()
	if err != nil {
		return nil, err
	}

	fi, _, _, err := hclread.NewEvaluationReadCloser(ctx, bytes.NewReader(cnt), filepath.Base(path.Path))
	if err != nil {
		return nil, err
	}

	pctx.Files[path.Path] = fi

	return pctx, nil
}

// Paths implements decoder.PathReader.
func (me *CacheSnapshotDecoder) Paths(ctx context.Context) []lang.Path {

	fles, err := me.snapshot.Symbols(context.Background(), false)
	if err != nil {
		return nil
	}

	paths := make([]lang.Path, 0, len(fles))
	for uri := range fles {
		paths = append(paths, lang.Path{Path: uri.Filename()})
	}

	return paths

}
