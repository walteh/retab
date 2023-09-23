// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package lsprpc implements a jsonrpc2.StreamServer that may be used to
// serve the LSP on a jsonrpc2 channel.
package lsprpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/walteh/retab/gen/gopls/event"
	"github.com/walteh/retab/gen/gopls/event/tag"
	"github.com/walteh/retab/gen/gopls/jsonrpc2"
	"github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/session"
)

// Unique identifiers for client/server.
var serverIndex int64

type ServerBuilder func(context.Context, *session.Session, protocol.ClientCloser) protocol.Server

// The StreamServer type is a jsonrpc2.StreamServer that handles incoming
// streams as a new LSP session, using a shared cache.
type StreamServer struct {
	// daemon controls whether or not to log new connections.
	daemon bool

	// serverForTest may be set to a test fake for testing.
	serverForTest protocol.Server

	serverBuilder ServerBuilder
}

// NewStreamServer creates a StreamServer using the shared cache. If
// withTelemetry is true, each session is instrumented with telemetry that
// records RPC statistics.
func NewStreamServer(isDaemon bool, builder ServerBuilder) *StreamServer {
	return &StreamServer{daemon: false, serverBuilder: builder}
}

func (s *StreamServer) Binder() *ServerBinder {
	newServer := func(ctx context.Context, client protocol.ClientCloser) protocol.Server {
		sess := session.NewSession()
		server := s.serverForTest
		if server == nil {
			server = s.serverBuilder(ctx, sess, client)
		}
		return server
	}
	return NewServerBinder(newServer)
}

// ServeStream implements the jsonrpc2.StreamServer interface, by handling
// incoming streams using a new lsp server.
func (s *StreamServer) ServeStream(ctx context.Context, conn jsonrpc2.Conn) error {
	client := protocol.ClientDispatcher(conn)
	sess := session.NewSession()
	server := s.serverForTest
	if server == nil {
		server = s.serverBuilder(ctx, sess, client)
	}
	// Clients may or may not send a shutdown message. Make sure the server is
	// shut down.
	// TODO(rFindley): this shutdown should perhaps be on a disconnected context.
	defer func() {
		if err := server.Shutdown(ctx); err != nil {
			event.Error(ctx, "error shutting down", err)
		}
	}()

	ctx = protocol.WithClient(ctx, client)
	conn.Go(ctx, protocol.Handlers(handshaker(sess, s.daemon, protocol.ServerHandler(server, jsonrpc2.MethodNotFound))))
	if s.daemon {
		log.Printf("Session %s: connected", sess.ID())
		defer log.Printf("Session %s: exited", sess.ID())
	}
	<-conn.Done()
	return conn.Err()
}

// A Forwarder is a jsonrpc2.StreamServer that handles an LSP stream by
// forwarding it to a remote. This is used when the gopls process started by
// the editor is in the `-remote` mode, which means it finds and connects to a
// separate gopls daemon. In these cases, we still want the forwarder gopls to
// be instrumented with telemetry, and want to be able to in some cases hijack
// the jsonrpc2 connection with the daemon.
type Forwarder struct {
	dialer *AutoDialer

	mu sync.Mutex
	// Hold on to the server connection so that we can redo the handshake if any
	// information changes.
	serverConn jsonrpc2.Conn
	serverID   string
}

// NewForwarder creates a new Forwarder, ready to forward connections to the
// remote server specified by rawAddr. If provided and rawAddr indicates an
// 'automatic' address (starting with 'auto;'), argFunc may be used to start a
// remote server for the auto-discovered address.
func NewForwarder(rawAddr string, argFunc func(network, address string) []string) (*Forwarder, error) {
	dialer, err := NewAutoDialer(rawAddr, argFunc)
	if err != nil {
		return nil, err
	}
	fwd := &Forwarder{
		dialer: dialer,
	}
	return fwd, nil
}

// // QueryServerState queries the server state of the current server.
// func QueryServerState(ctx context.Context, addr string) (*ServerState, error) {
// 	serverConn, err := dialRemote(ctx, addr)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var state ServerState
// 	if err := protocol.Call(ctx, serverConn, sessionsMethod, nil, &state); err != nil {
// 		return nil, fmt.Errorf("querying server state: %w", err)
// 	}
// 	return &state, nil
// }

// dialRemote is used for making calls into the gopls daemon. addr should be a
// URL, possibly on the synthetic 'auto' network (e.g. tcp://..., unix://...,
// or auto://...).
func dialRemote(ctx context.Context, addr string) (jsonrpc2.Conn, error) {
	network, address := ParseAddr(addr)
	if network == AutoNetwork {
		gp, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("getting gopls path: %w", err)
		}
		network, address = autoNetworkAddress(gp, address)
	}
	netConn, err := net.DialTimeout(network, address, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("dialing remote: %w", err)
	}
	serverConn := jsonrpc2.NewConn(jsonrpc2.NewHeaderStream(netConn))
	serverConn.Go(ctx, jsonrpc2.MethodNotFound)
	return serverConn, nil
}

// func ExecuteCommand(ctx context.Context, addr string, id string, request, result interface{}) error {
// 	serverConn, err := dialRemote(ctx, addr)
// 	if err != nil {
// 		return err
// 	}
// 	args, err := command.MarshalArgs(request)
// 	if err != nil {
// 		return err
// 	}
// 	params := protocol.ExecuteCommandParams{
// 		Command:   id,
// 		Arguments: args,
// 	}
// 	return protocol.Call(ctx, serverConn, "workspace/executeCommand", params, result)
// }

// ServeStream dials the forwarder remote and binds the remote to serve the LSP
// on the incoming stream.
func (f *Forwarder) ServeStream(ctx context.Context, clientConn jsonrpc2.Conn) error {
	client := protocol.ClientDispatcher(clientConn)

	netConn, err := f.dialer.dialNet(ctx)
	if err != nil {
		return fmt.Errorf("forwarder: connecting to remote: %w", err)
	}
	serverConn := jsonrpc2.NewConn(jsonrpc2.NewHeaderStream(netConn))
	server := protocol.ServerDispatcher(serverConn)

	// Forward between connections.
	serverConn.Go(ctx,
		protocol.Handlers(
			protocol.ClientHandler(client,
				jsonrpc2.MethodNotFound)))

	// Don't run the clientConn yet, so that we can complete the handshake before
	// processing any client messages.

	// Do a handshake with the server instance to exchange debug information.
	index := atomic.AddInt64(&serverIndex, 1)
	f.mu.Lock()
	f.serverConn = serverConn
	f.serverID = strconv.FormatInt(index, 10)
	f.mu.Unlock()
	clientConn.Go(ctx,
		protocol.Handlers(
			f.handler(
				protocol.ServerHandler(server,
					jsonrpc2.MethodNotFound))))

	select {
	case <-serverConn.Done():
		clientConn.Close()
	case <-clientConn.Done():
		serverConn.Close()
	}

	err = nil
	if serverConn.Err() != nil {
		err = fmt.Errorf("remote disconnected: %v", serverConn.Err())
	} else if clientConn.Err() != nil {
		err = fmt.Errorf("client disconnected: %v", clientConn.Err())
	}
	event.Log(ctx, fmt.Sprintf("forwarder: exited with error: %v", err))
	return err
}

func ConnectToRemote(ctx context.Context, addr string) (net.Conn, error) {
	dialer, err := NewAutoDialer(addr, nil)
	if err != nil {
		return nil, err
	}
	return dialer.dialNet(ctx)
}

// handler intercepts messages to the daemon to enrich them with local
// information.
func (f *Forwarder) handler(handler jsonrpc2.Handler) jsonrpc2.Handler {
	return func(ctx context.Context, reply jsonrpc2.Replier, r jsonrpc2.Request) error {
		// Intercept certain messages to add special handling.
		switch r.Method() {
		case "initialize":

		}
		// The gopls workspace environment defaults to the process environment in
		// which gopls daemon was started. To avoid discrepancies in Go environment
		// between the editor and daemon, inject any unset variables in `go env`
		// into the options sent by initialize.
		//
		// See also golang.org/issue/37830.
		return handler(ctx, reply, r)
	}
}

// A handshakeRequest identifies a client to the LSP server.
type handshakeRequest struct {
	// ServerID is the ID of the server on the client. This should usually be 0.
	ServerID string `json:"serverID"`
	// Logfile is the location of the clients log file.
	Logfile string `json:"logfile"`
	// DebugAddr is the client debug address.
	DebugAddr string `json:"debugAddr"`
	// GoplsPath is the path to the Gopls binary running the current client
	// process.
	GoplsPath string `json:"goplsPath"`
}

// A handshakeResponse is returned by the LSP server to tell the LSP client
// information about its session.
type handshakeResponse struct {
	// SessionID is the server session associated with the client.
	SessionID string `json:"sessionID"`
	// Logfile is the location of the server logs.
	Logfile string `json:"logfile"`
	// DebugAddr is the server debug address.
	DebugAddr string `json:"debugAddr"`
	// GoplsPath is the path to the Gopls binary running the current server
	// process.
	GoplsPath string `json:"goplsPath"`
}

// ClientSession identifies a current client LSP session on the server. Note
// that it looks similar to handshakeResposne, but in fact 'Logfile' and
// 'DebugAddr' now refer to the client.
type ClientSession struct {
	SessionID string `json:"sessionID"`
	Logfile   string `json:"logfile"`
	DebugAddr string `json:"debugAddr"`
}

// ServerState holds information about the gopls daemon process, including its
// debug information and debug information of all of its current connected
// clients.
type ServerState struct {
	Logfile         string          `json:"logfile"`
	CurrentClientID string          `json:"currentClientID"`
	Clients         []ClientSession `json:"clients"`
}

const (
	handshakeMethod = "gopls/handshake"
	sessionsMethod  = "gopls/sessions"
)

func handshaker(sess *session.Session, logHandshakes bool, handler jsonrpc2.Handler) jsonrpc2.Handler {
	return func(ctx context.Context, reply jsonrpc2.Replier, r jsonrpc2.Request) error {
		switch r.Method() {
		case handshakeMethod:
			// We log.Printf in this handler, rather than event.Log when we want logs
			// to go to the daemon log rather than being reflected back to the
			// client.
			var req handshakeRequest
			if err := json.Unmarshal(r.Params(), &req); err != nil {
				if logHandshakes {
					log.Printf("Error processing handshake for session %s: %v", sess.ID(), err)
				}
				sendError(ctx, reply, err)
				return nil
			}
			if logHandshakes {
				log.Printf("Session %s: got handshake. Logfile: %q, Debug addr: %q", sess.ID(), req.Logfile, req.DebugAddr)
			}
			event.Log(ctx, "Handshake session update",
				tag.DebugAddress.Of(req.DebugAddr),
				tag.Logfile.Of(req.Logfile),
				tag.ServerID.Of(req.ServerID),
				tag.GoplsPath.Of(req.GoplsPath),
			)
			resp := handshakeResponse{
				SessionID: sess.ID(),
			}
			return reply(ctx, resp, nil)

		case sessionsMethod:
			resp := ServerState{
				CurrentClientID: sess.ID(),
			}
			return reply(ctx, resp, nil)
		}
		return handler(ctx, reply, r)
	}
}

func sendError(ctx context.Context, reply jsonrpc2.Replier, err error) {
	err = fmt.Errorf("%v: %w", err, jsonrpc2.ErrParse)
	if err := reply(ctx, nil, err); err != nil {
		event.Error(ctx, "", err)
	}
}

// ParseAddr parses the address of a gopls remote.
// TODO(rFindley): further document this syntax, and allow URI-style remote
// addresses such as "auto://...".
func ParseAddr(listen string) (network string, address string) {
	// Allow passing just -remote=auto, as a shorthand for using automatic remote
	// resolution.
	if listen == AutoNetwork {
		return AutoNetwork, ""
	}
	if parts := strings.SplitN(listen, ";", 2); len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "tcp", listen
}
