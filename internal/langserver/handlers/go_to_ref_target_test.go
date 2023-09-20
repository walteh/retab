// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"fmt"
	"testing"

	"github.com/walteh/retab/internal/langserver"
	"github.com/walteh/retab/internal/state"
)

func TestDefinition_basic(t *testing.T) {
	tmpDir := TempDir(t)

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
				"textDocument": {
					"declaration": {
						"linkSupport": true
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
			"languageId": "terraform",
			"text": `+fmt.Sprintf("%q",
			`resource "test_resource_2" "foo" {
    setting {
        name  = "foo"
        value = "bar"
    }
}

output "foo" {
    value = test_resource_2.foo.setting
}`)+`,
			"uri": "%s/main.tf"
		}
	}`, tmpDir.URI)})
	waitForAllJobs(t, ss)

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method: "textDocument/declaration",
		ReqParams: fmt.Sprintf(`{
			"textDocument": {
				"uri": "%s/main.tf"
			},
			"position": {
				"line": 8,
				"character": 35
			}
		}`, tmpDir.URI)}, fmt.Sprintf(`{
			"jsonrpc": "2.0",
			"id": 3,
			"result": [
				{
					"originSelectionRange": {
						"start": {
							"line": 8,
							"character": 12
						},
						"end": {
							"line": 8,
							"character": 39
						}
					},
					"targetUri": "%s/main.tf",
					"targetRange": {
						"start": {
							"line": 1,
							"character": 4
						},
						"end": {
							"line": 4,
							"character": 5
						}
					},
					"targetSelectionRange": {
						"start": {
							"line": 1,
							"character": 4
						},
						"end": {
							"line": 4,
							"character": 5
						}
					}
				}
			]
		}`, tmpDir.URI))
}
