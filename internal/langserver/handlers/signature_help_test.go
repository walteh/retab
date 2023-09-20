// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"encoding/json"
	"fmt"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/walteh/retab/internal/langserver"
	"github.com/walteh/retab/internal/langserver/session"
	"github.com/walteh/retab/internal/state"
)

func TestSignatureHelp_withoutInitialization(t *testing.T) {
	ls := langserver.NewLangServerMock(t, NewMockSession(nil))
	stop := ls.Start(t)
	defer stop()

	ls.CallAndExpectError(t, &langserver.CallRequest{
		Method: "textDocument/signatureHelp",
		ReqParams: fmt.Sprintf(`{
			"textDocument": {
				"uri": "%s/main.tf"
			},
			"position": {
				"character": 0,
				"line": 1
			},
			"context": {
				"isRetrigger": false,
				"triggerCharacter": "(",
				"triggerKind": 2
			}
		}`, TempDir(t).URI)}, session.SessionNotInitialized.Err())
}

func TestSignatureHelp_withValidData(t *testing.T) {
	tmpDir := TempDir(t)
	InitPluginCache(t, tmpDir.Path())

	var testSchema tfjson.ProviderSchemas
	err := json.Unmarshal([]byte(testModuleSchemaOutput), &testSchema)
	if err != nil {
		t.Fatal(err)
	}

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
			"languageId": "terraform",
			"text": "variable \"name\" {\n  default = file(\"~/foo\")\n}",
			"uri": "%s/main.tf"
		}
	}`, TempDir(t).URI)})
	waitForAllJobs(t, ss)

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method: "textDocument/signatureHelp",
		ReqParams: fmt.Sprintf(`{
			"textDocument": {
				"uri": "%s/main.tf"
			},
			"position": {
				"character": 16,
				"line": 1
			},
			"context": {
				"isRetrigger": false,
				"triggerCharacter": "(",
				"triggerKind": 2
			}
		}`, TempDir(t).URI)}, `{
			"jsonrpc": "2.0",
			"id": 3,
			"result": {
				"signatures": [{
					"label": "file(path string) string",
					"documentation": "file reads the contents of a file at the given path and returns them as a string.",
					"parameters": [{"label": "path"}]
				}]
			}
		}`)
}
