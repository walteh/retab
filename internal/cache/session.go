package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/spf13/afero"
	"github.com/walteh/retab/internal/lsp/filesystem"
	"github.com/walteh/retab/internal/lsp/utm"
)

type Cache struct {
	id string

	decoder *decoder.Decoder
	fs      afero.Fs
}

func (s *Cache) Decoder() *decoder.Decoder {
	return s.decoder
}

func (s *Cache) Fs() afero.Fs {
	return s.fs
}

func (s *Cache) ID() string {
	return s.id
}

func NewCache(ctx context.Context) *Cache {
	aferoFs := afero.NewMemMapFs()
	fls := filesystem.NewFilesystem(aferoFs)
	dec := decoder.NewDecoder(fls)
	dCtx := decoder.NewDecoderContext()
	dCtx.UtmSource = utm.UtmSource
	dCtx.UtmMedium = utm.UtmMedium(ctx)
	dCtx.UseUtmContent = true
	return &Cache{
		id:      strconv.FormatInt(time.Now().UnixNano(), 10),
		decoder: dec,
		fs:      aferoFs,
	}
}

type Session struct {
	id    string
	cache *Cache
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) Cache() *Cache {
	return s.cache
}

func (s *Session) Views() []*View {
	return []*View{}
}

func (s *Session) Overlays() []*Overlay {
	return []*Overlay{}
}

func NewSession(ctx context.Context, cache *Cache) *Session {
	return &Session{
		id:    strconv.FormatInt(time.Now().UnixNano(), 10),
		cache: cache,
	}
}
