// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"
	"io"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/spf13/afero"
	"github.com/walteh/retab/internal/lsp/filesystem"
	"github.com/walteh/retab/internal/lsp/langserver/session"
)

type MockSessionInput struct {
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

	svc := &service{
		logger:      testLogger(),
		srvCtx:      srvCtx,
		sessCtx:     sessCtx,
		stopSession: ms.stop,
		fs:          nil,
	}

	return svc
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
	ms := &mockSession{
		stopCalledMu: &sync.RWMutex{},
	}
	srvCtx := context.Background()
	sessCtx, stopSession := context.WithCancel(srvCtx)
	ms.stopFunc = stopSession

	fs := filesystem.NewFilesystem(afero.NewMemMapFs())

	svc := &service{
		logger:      testLogger(),
		srvCtx:      srvCtx,
		sessCtx:     sessCtx,
		stopSession: ms.stop,
		decoder:     decoder.NewDecoder(fs),
		fs:          fs,
	}

	return func(ctx context.Context) session.Session {
		return svc
	}
}
