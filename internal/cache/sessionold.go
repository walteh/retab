package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/spf13/afero"
)

type CacheOLD struct {
	id string

	decoder *decoder.Decoder
	fs      afero.Fs
}

func (s *CacheOLD) Decoder() *decoder.Decoder {
	return s.decoder
}

func (s *CacheOLD) Fs() afero.Fs {
	return s.fs
}

func (s *CacheOLD) ID() string {
	return s.id
}

// func NewCacheOLD(ctx context.Context) *CacheOLD {
// 	aferoFs := afero.NewMemMapFs()
// 	fls := filesystem.NewFilesystem(aferoFs)
// 	dec := decoder.NewDecoder(fls)
// 	dCtx := decoder.NewDecoderContext()
// 	dCtx.UtmSource = utm.UtmSource
// 	dCtx.UtmMedium = utm.UtmMedium(ctx)
// 	dCtx.UseUtmContent = true
// 	return &CacheOLD{
// 		id:      strconv.FormatInt(time.Now().UnixNano(), 10),
// 		decoder: dec,
// 		fs:      aferoFs,
// 	}
// }

type SessionOLD struct {
	id    string
	cache *CacheOLD
}

func (s *SessionOLD) ID() string {
	return s.id
}

func (s *SessionOLD) CacheOLD() *CacheOLD {
	return s.cache
}

func (s *SessionOLD) Views() []*View {
	return []*View{}
}

func (s *SessionOLD) Overlays() []*Overlay {
	return []*Overlay{}
}

func NewSessionOLD(ctx context.Context, cache *CacheOLD) *SessionOLD {
	return &SessionOLD{
		id:    strconv.FormatInt(time.Now().UnixNano(), 10),
		cache: cache,
	}
}
