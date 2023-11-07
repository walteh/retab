package lsp

import (
	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
	"github.com/walteh/retab/version"

	// Must include a backend implementation
	// See CommonLog for other options: https://github.com/tliron/commonlog
	_ "github.com/tliron/commonlog/simple"
)

const lsName = "retab"

type Handler struct {
	protocol.Handler
	files map[string]string
}

var handler Handler

func NewServer() *server.Server {
	// This increases logging verbosity (optional)
	commonlog.Configure(1, nil)

	handler = Handler{
		Handler: protocol.Handler{
			Initialize:  initialize,
			Initialized: initialized,
			Shutdown:    shutdown,
			SetTrace:    setTrace,
			TextDocumentDidOpen: func(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
				handler.files[params.TextDocument.URI] = params.TextDocument.Text
				return nil
			},
			TextDocumentDidChange: func(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
				return nil
			},
		},
	}

	return server.NewServer(&handler, lsName, false)
}

func initialize(_ *glsp.Context, _ *protocol.InitializeParams) (any, error) {
	capabilities := handler.CreateServerCapabilities()

	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    lsName,
			Version: &version.Version,
		},
	}, nil
}

func initialized(_ *glsp.Context, _ *protocol.InitializedParams) error {
	return nil
}

func shutdown(_ *glsp.Context) error {
	protocol.SetTraceValue(protocol.TraceValueOff)
	return nil
}

func setTrace(_ *glsp.Context, params *protocol.SetTraceParams) error {
	protocol.SetTraceValue(params.Value)
	return nil
}
