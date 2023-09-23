// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package lsp implements LSP for gopls.
package server

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/hashicorp/hcl/v2"
	"github.com/walteh/retab/gen/gopls/event"
	"github.com/walteh/retab/gen/gopls/jsonrpc2"
	"github.com/walteh/retab/gen/gopls/progress"
	"github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/gen/gopls/span"
	"github.com/walteh/retab/internal/session"
)

const concurrentAnalyses = 1

// NewServer creates an LSP server and binds it to handle incoming client
// messages on the supplied stream.
func NewServer(ctx context.Context, sess *session.Session, client protocol.ClientCloser) *Server {
	return &Server{
		diagnostics:         map[span.URI]*hcl.Diagnostics{},
		watchedGlobPatterns: nil, // empty
		changedFiles:        make(map[span.URI]struct{}),
		session:             sess,
		client:              client,
		diagnosticsSema:     make(chan struct{}, concurrentAnalyses),
		progress:            progress.NewTracker(client),
	}
}

func NewProtocolServer(ctx context.Context, sess *session.Session, client protocol.ClientCloser) protocol.Server {
	return NewServer(ctx, sess, client)
}

type serverState int

const (
	serverCreated      = serverState(iota)
	serverInitializing // set once the server has received "initialize" request
	serverInitialized  // set once the server has received "initialized" request
	serverShutDown
)

func (s serverState) String() string {
	switch s {
	case serverCreated:
		return "created"
	case serverInitializing:
		return "initializing"
	case serverInitialized:
		return "initialized"
	case serverShutDown:
		return "shutDown"
	}
	return fmt.Sprintf("(unknown state: %d)", int(s))
}

// Server implements the protocol.Server interface.
type Server struct {
	client protocol.ClientCloser

	stateMu sync.Mutex
	state   serverState
	// notifications generated before serverInitialized
	notifications []*protocol.ShowMessageParams

	session *session.Session

	tempDir string

	// changedFiles tracks files for which there has been a textDocument/didChange.
	changedFilesMu sync.Mutex
	changedFiles   map[span.URI]struct{}

	// folders is only valid between initialize and initialized, and holds the
	// set of folders to build views for when we are ready
	pendingFolders []protocol.WorkspaceFolder

	// watchedGlobPatterns is the set of glob patterns that we have requested
	// the client watch on disk. It will be updated as the set of directories
	// that the server should watch changes.
	// The map field may be reassigned but the map is immutable.
	watchedGlobPatternsMu  sync.Mutex
	watchedGlobPatterns    map[string]struct{}
	watchRegistrationCount int

	diagnosticsMu sync.Mutex
	diagnostics   map[span.URI]*hcl.Diagnostics

	// diagnosticsSema limits the concurrency of diagnostics runs, which can be
	// expensive.
	diagnosticsSema chan struct{}

	progress *progress.Tracker

	// When the workspace fails to load, we show its status through a progress
	// report with an error message.
	criticalErrorStatusMu sync.Mutex
	criticalErrorStatus   *progress.WorkDone

	// Track an ongoing CPU profile created with the StartProfile command and
	// terminated with the StopProfile command.
	ongoingProfileMu sync.Mutex
	ongoingProfile   *os.File // if non-nil, an ongoing profile is writing to this file

}

func (s *Server) workDoneProgressCancel(ctx context.Context, params *protocol.WorkDoneProgressCancelParams) error {
	ctx, done := event.Start(ctx, "lsp.Server.workDoneProgressCancel")
	defer done()

	return s.progress.Cancel(params.Token)
}

func (s *Server) nonstandardRequest(ctx context.Context, method string, params interface{}) (interface{}, error) {
	ctx, done := event.Start(ctx, "lsp.Server.nonstandardRequest")
	defer done()

	switch method {

	}
	return nil, notImplemented(method)
}

func notImplemented(method string) error {
	return fmt.Errorf("%w: %q not yet implemented", jsonrpc2.ErrMethodNotFound, method)
}
