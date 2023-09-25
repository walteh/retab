package lsp

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/walteh/retab/gen/gopls/event"
	"github.com/walteh/retab/gen/gopls/event/core"
	"github.com/walteh/retab/gen/gopls/event/label"
	"github.com/walteh/retab/gen/gopls/fakenet"
	"github.com/walteh/retab/gen/gopls/jsonrpc2"
	"github.com/walteh/retab/gen/gopls/protocol"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

// func Test2(t *testing.T) {
// 	t.Run("test", func(t *testing.T) {
// 		listener, _ := net.Listen("tcp", ":8080")
// 		conn, _ := listener.Accept()
// 		defer conn.Close()

// 		// Create JSON RPC connection
// 		jconn := jsonrpc2.NewConn(jsonrpc2.NewRawStream(conn))

// 		// Start the server
// 		go StartServer(jconn)

// 		// Start the client
// 		client := jsonrpc2.NewClient(jconn)
// 	})
// }

func TestFull(t *testing.T) {

	t.Run("test", func(t *testing.T) {
		ctx := context.Background()

		fls := afero.NewMemMapFs()

		ss := NewServer(fls)

		event.SetExporter(func(ctx context.Context, e core.Event, m label.Map) context.Context {
			// pp.Println(e)
			return ctx
		})

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

		conn.Go(ctx, NewHandler()

		// event.SetExporter(func(ctx context.Context, e core.Event, m label.Map) context.Context {
		// 	// pp.Println(e)
		// 	return ctx
		// })

		// // conn := fakenet.NewConn("stdio", os.Stdin, os.Stdout)

		// fls := afero.NewMemMapFs()

		// ss := NewServer(fls)

		// // srv := NewServer(fls, protocol.Client))

		// handlez := protocol.ServerHandlerV2(ss)

		// abc := NewHandler(handlez)

		// srvr := abc.AsServer(ctx)

		// in := strings.NewReader("Content-Length: 81\r\n\r\n{\"jsonrpc\": \"2.0\", \"id\": 1, \"method\": \"initialize\", \"params\": {\"processId\": null}}")
		// or, ow := io.Pipe()

		// srvr = srvr.Start(channel.LSP(in, ow))

		// go func() {
		// 	for {
		// 		outStr, err := io.ReadAll(or)
		// 		if err != nil {
		// 			return
		// 		}

		// 		t.Log("out -------- ", string(outStr))
		// 	}
		// }()

		// Sending initialize request to the LSP server

		_ = srvr.Wait()

		// ln, err := net.Listen("tcp", ":8090")
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// for {
		// 	conn, err := ln.Accept()
		// 	if err != nil {
		// 		log.Println("Accept error:", err)
		// 		continue
		// 	}

		// 	go srvr.Start(channel.LSP(conn, conn))
		// }

		// srvr = srvr.Start(channel.LSP(os.Stdin, os.Stdout))

		// stat := srvr.WaitStatus()

		// pp.Println("yo", stat)
	})
}
