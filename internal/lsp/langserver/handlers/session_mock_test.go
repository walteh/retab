// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/creachadair/jrpc2/handler"
	"github.com/walteh/retab/internal/lsp/langserver/session"
	"github.com/walteh/retab/internal/lsp/state"
	"github.com/walteh/retab/internal/lsp/walker"
)

type MockSessionInput struct {
	AdditionalHandlers map[string]handler.Func
	StateStore         *state.StateStore
	WalkerCollector    *walker.WalkerCollector
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
	var walkerCollector *walker.WalkerCollector
	if ms.mockInput != nil {
		stateStore = ms.mockInput.StateStore
		walkerCollector = ms.mockInput.WalkerCollector
		handlers = ms.mockInput.AdditionalHandlers
	}

	svc := &service{
		logger:             testLogger(),
		srvCtx:             srvCtx,
		sessCtx:            sessCtx,
		stopSession:        ms.stop,
		additionalHandlers: handlers,
		stateStore:         stateStore,
		walkerCollector:    walkerCollector,
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

	return log.New(ioutil.Discard, "", 0)
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
