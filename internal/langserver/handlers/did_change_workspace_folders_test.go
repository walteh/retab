// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"fmt"
	"testing"

	"github.com/walteh/retab/internal/langserver"
	"github.com/walteh/retab/internal/state"
)

func TestDidChangeWorkspaceFolders(t *testing.T) {
	rootDir := TempDir(t)

	ss, err := state.NewStateStore()
	if err != nil {
		t.Fatal(err)
	}

	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{
		StateStore: ss,
	}))
	stop := ls.Start(t)
	defer stop()

	ls.Call(t, &langserver.CallRequest{
		Method: "initialize",
		ReqParams: fmt.Sprintf(`{
		"capabilities": {},
		"rootUri": %q,
		"processId": 12345,
		"workspaceFolders": [
			{
				"uri": %q,
				"name": "first"
			}
		]
	}`, rootDir.URI, rootDir.URI)})

	ls.Notify(t, &langserver.CallRequest{
		Method:    "initialized",
		ReqParams: "{}",
	})
	ls.Call(t, &langserver.CallRequest{
		Method: "workspace/didChangeWorkspaceFolders",
		ReqParams: fmt.Sprintf(`{
		"event": {
			"added": [
				{"uri": %q, "name": "second"}
			],
			"removed": [
				{"uri": %q, "name": "first"}
			]
		}
	}`, rootDir.URI, rootDir.URI)})

}
