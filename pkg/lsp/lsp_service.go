package lsp

import (
	"context"
	"encoding/json"
	"log"

	"github.com/creachadair/jrpc2"
	"github.com/walteh/retab/gen/gopls/event"
	jsonrpc2 "github.com/walteh/retab/gen/gopls/jsonrpc2_v2"
)

type handler struct {
	jsonrpc2.Handler
}

var _ jrpc2.Assigner = (*handler)(nil)

func NewHandler(h jsonrpc2.Handler) *handler {
	return &handler{h}
}

func (svc *handler) Assign(ctx context.Context, method string) jrpc2.Handler {
	return func(ctx context.Context, r *jrpc2.Request) (any, error) {
		return svc.Handle(ctx, &jsonrpc2.Request{
			ID:     jsonrpc2.StringID(r.ID()),
			Method: r.Method(),
			Params: json.RawMessage(r.ParamString()),
		})
	}
}

func (svc *handler) AsServer(ctx context.Context) *jrpc2.Server {
	return jrpc2.NewServer(svc, &jrpc2.ServerOptions{
		AllowPush:   true,
		Concurrency: 1,
		NewContext: func() context.Context {
			return ctx
		},
		Logger: func(text string) {
			log.Println(text)
			event.Log(ctx, text)
		},
	})
}
