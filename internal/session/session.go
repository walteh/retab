package session

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/spf13/afero"
	"github.com/walteh/retab/internal/lsp/filesystem"
	"github.com/walteh/retab/internal/lsp/utm"
)

type Session struct {
	id string

	decoder *decoder.Decoder
	fs      afero.Fs
}

func (s *Session) Decoder() *decoder.Decoder {
	return s.decoder
}

func (s *Session) Fs() afero.Fs {
	return s.fs
}

func (s *Session) ID() string {
	return s.id
}

func NewSession(ctx context.Context) *Session {
	aferoFs := afero.NewMemMapFs()
	fls := filesystem.NewFilesystem(aferoFs)
	dec := decoder.NewDecoder(fls)
	dCtx := decoder.NewDecoderContext()
	dCtx.UtmSource = utm.UtmSource
	dCtx.UtmMedium = utm.UtmMedium(ctx)
	dCtx.UseUtmContent = true
	return &Session{
		id:      strconv.FormatInt(time.Now().UnixNano(), 10),
		decoder: dec,
		fs:      aferoFs,
	}
}
