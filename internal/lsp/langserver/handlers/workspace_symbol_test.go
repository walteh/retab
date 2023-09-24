// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"fmt"
	"testing"

	"github.com/walteh/retab/internal/lsp/langserver"
	"github.com/walteh/retab/internal/lsp/state"
)

func TestLangServer_workspace_symbol_basic(t *testing.T) {
	tmpDir := TempDir(t)
	InitPluginCache(t, tmpDir.Path())

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
		"capabilities": {
			"workspace": {
				"symbol": {
					"symbolKind": {
						"valueSet": [
							1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
							11, 12, 13, 14, 15, 16, 17, 18,
							19, 20, 21, 22, 23, 24, 25, 26
						]
					},
					"tagSupport": {
						"valueSet": [ 1 ]
					}
				}
			}
		},
		"rootUri": %q,
		"processId": 12345
	}`, tmpDir.URI)})

	ls.Notify(t, &langserver.CallRequest{
		Method:    "initialized",
		ReqParams: "{}",
	})
	ls.Call(t, &langserver.CallRequest{
		Method: "textDocument/didOpen",
		ReqParams: fmt.Sprintf(`{
		"textDocument": {
			"version": 0,
			"languageId": "retab",
			"text": "provider \"github\" {}",
			"uri": "%s/first.tf"
		}
	}`, tmpDir.URI)})
	ls.Call(t, &langserver.CallRequest{
		Method: "textDocument/didOpen",
		ReqParams: fmt.Sprintf(`{
		"textDocument": {
			"version": 0,
			"languageId": "retab",
			"text": "provider \"google\" {}",
			"uri": "%s/second.tf"
		}
	}`, tmpDir.URI)})
	ls.Call(t, &langserver.CallRequest{
		Method: "textDocument/didOpen",
		ReqParams: fmt.Sprintf(`{
		"textDocument": {
			"version": 0,
			"languageId": "retab",
			"text": "myblock \"custom\" {}",
			"uri": "%s/blah/third.tf"
		}
	}`, tmpDir.URI)})

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method: "workspace/symbol",
		ReqParams: `{
		"query": ""
	}`}, fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"id": 5,
		"result": [
			{
				"location": {
					"uri": "%s/first.tf",
					"range": {
						"start": {"line": 0, "character": 0},
						"end": {"line": 0, "character": 20}
					}
				},
				"name": "provider \"github\"",
				"kind": 5
			},
			{
				"location": {
					"uri": "%s/second.tf",
					"range": {
						"start": {"line": 0, "character": 0},
						"end": {"line": 0, "character": 20}
					}
				},
				"name": "provider \"google\"",
				"kind": 5
			},
			{
				"location": {
					"uri": "%s/blah/third.tf",
					"range": {
						"start": {"line": 0, "character": 0},
						"end": {"line": 0, "character": 19}
					}
				},
				"name": "myblock \"custom\"",
				"kind": 5
			}
		]
	}`, tmpDir.URI, tmpDir.URI, tmpDir.URI))

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method: "workspace/symbol",
		ReqParams: `{
		"query": "myb"
	}`}, fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"id": 6,
		"result": [
			{
				"location": {
					"uri": "%s/blah/third.tf",
					"range": {
						"start": {"line": 0, "character": 0},
						"end": {"line": 0, "character": 19}
					}
				},
				"name": "myblock \"custom\"",
				"kind": 5
			}
		]
	}`, tmpDir.URI))
}
