package lsp

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/k0kubun/pp/v3"
	"github.com/walteh/retab/gen/gopls/fakenet"
	"github.com/walteh/retab/gen/gopls/jsonrpc2"
	"github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/debug"
)

func TestFull(t *testing.T) {

	ctx := debug.WithInstance(context.Background(), "./de.bug", "serve")

	t.Run("test", func(t *testing.T) {

		srv := NewServe()

		stream := jsonrpc2.NewHeaderStream(fakenet.NewConn("stdio", os.Stdin, os.Stdout))
		// if s.Trace {
		// 	stream = protocol.LoggingStream(stream, di.LogWriter)
		// }
		// log.Printf("Gopls daemon: serving on stdin/stdout...")
		conn := jsonrpc2.NewConn(stream)

		err := protocol.ClientDispatcher(conn).ShowMessage(ctx, &protocol.ShowMessageParams{
			Type:    protocol.Info,
			Message: "Hello World",
		})
		if err != nil {
			t.Fatal(err)
		}

		go func() {
			if err := srv.Run(ctx, conn); err != nil {
				t.Fatal(err)
			}
		}()

		resp, err := http.Post("http://localhost:8090", "application/json", nil)
		if err != nil {
			t.Fatal(err)
		}

		pp.Println(resp)

		<-ctx.Done()

	})
}
