// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"fmt"
	"testing"

	"github.com/creachadair/jrpc2"
	"github.com/walteh/retab/internal/lsp/langserver"
)

func TestShutdown_twice(t *testing.T) {
	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{}))
	stop := ls.Start(t)
	defer stop()

	ls.Call(t, &langserver.CallRequest{
		Method: "initialize",
		ReqParams: fmt.Sprintf(`{
	    "capabilities": {},
	    "rootUri": %q,
	    "processId": 12345
	}`, TempDir(t).URI)})
	ls.Call(t, &langserver.CallRequest{
		Method: "shutdown", ReqParams: `{}`})

	ls.CallAndExpectError(t, &langserver.CallRequest{
		Method: "shutdown", ReqParams: `{}`},
		jrpc2.InvalidRequest.Err())
}
