// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/creachadair/jrpc2/handler"
	"github.com/walteh/retab/internal/lsp/langserver/session"
	"github.com/walteh/retab/internal/lsp/state"
)

type MockSessionInput struct {
	AdditionalHandlers map[string]handler.Func
	StateStore         *state.StateStore
}

type mockSession struct {
	mockInput    *MockSessionInput
	stopFunc     func()
	stopCalled   bool
	stopCalledMu *sync.RWMutex
}

func (ms *mockSession) new(srvCtx context.Context) session.Session {
	sessCtx, stopSession := context.WithCancel(srvCtx)
	ms.stopFunc = stopSession

	var handlers map[string]handler.Func
	var stateStore *state.StateStore
	if ms.mockInput != nil {
		stateStore = ms.mockInput.StateStore
		handlers = ms.mockInput.AdditionalHandlers
	}

	svc := &service{
		logger:             testLogger(),
		srvCtx:             srvCtx,
		sessCtx:            sessCtx,
		stopSession:        ms.stop,
		additionalHandlers: handlers,
		stateStore:         stateStore,
	}

	return svc
}

func defaultRegistryServer() *httptest.Server {
	return httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unexpected Registry API request", 500)
	}))
}

func testLogger() *log.Logger {
	if testing.Verbose() {
		return log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	}

	return log.New(io.Discard, "", 0)
}

func (ms *mockSession) stop() {

	ms.stopCalledMu.Lock()
	defer ms.stopCalledMu.Unlock()

	ms.stopFunc()
	ms.stopCalled = true
}

func (ms *mockSession) StopFuncCalled() bool {
	ms.stopCalledMu.RLock()
	defer ms.stopCalledMu.RUnlock()

	return ms.stopCalled
}

func newMockSession(input *MockSessionInput) *mockSession {
	return &mockSession{
		mockInput:    input,
		stopCalledMu: &sync.RWMutex{},
	}
}

func NewMockSession(input *MockSessionInput) session.SessionFactory {
	return newMockSession(input).new
}
