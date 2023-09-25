package lsp

import (
	"path/filepath"

	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/spf13/afero"
	"github.com/walteh/retab/gen/gopls/protocol"
)

var _ protocol.Server = (*Server)(nil)

type Server struct {
	imfs afero.Fs
}

func NewServer(imfs afero.Fs) *Server {
	return &Server{
		imfs: imfs,
	}
}

func (s *Server) DecoderForURI(filename protocol.TextDocumentIdentifier) *decoder.Decoder {
	pr := NewPathReader(s.imfs, filepath.Dir(string(filename.URI)))

	return decoder.NewDecoder(pr)
}
