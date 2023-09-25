package lsp

import (
	"context"
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

		con := fakenet.NewConn("stdio", os.Stdin, os.Stdout)

		stream := jsonrpc2.NewHeaderStream(con)
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
				t.Log(err)
			}
		}()

		conn.Go()

		pp.Println(resp)

		<-ctx.Done()

	})
}
