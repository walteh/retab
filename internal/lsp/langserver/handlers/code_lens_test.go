// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/walteh/retab/internal/lsp/document"
	"github.com/walteh/retab/internal/lsp/langserver"
	"github.com/walteh/retab/internal/lsp/langserver/session"
	"github.com/walteh/retab/internal/lsp/state"
)

func TestCodeLens_withoutInitialization(t *testing.T) {
	ls := langserver.NewLangServerMock(t, NewMockSession(nil))
	stop := ls.Start(t)
	defer stop()

	ls.CallAndExpectError(t, &langserver.CallRequest{
		Method: "textDocument/codeLens",
		ReqParams: fmt.Sprintf(`{
			"textDocument": {
				"uri": "%s/main.tf"
			}
		}`, TempDir(t).URI)}, session.SessionNotInitialized.Err())
}

func TestCodeLens_withoutOptIn(t *testing.T) {
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
			"languageId": "retab",
			"text": "provider \"test\" {\n\n}\n",
			"uri": "%s/main.tf"
		}
	}`, tmpDir.URI)})

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method: "textDocument/codeLens",
		ReqParams: fmt.Sprintf(`{
			"textDocument": {
				"uri": "%s/main.tf"
			}
		}`, TempDir(t).URI),
	}, `{
				"jsonrpc": "2.0",
				"id": 3,
				"result": []
	}`)
}

func TestCodeLens_referenceCount(t *testing.T) {
	tmpDir := TempDir(t)
	InitPluginCache(t, tmpDir.Path())

	// var testSchema tfjson.ProviderSchemas
	// err := json.Unmarshal([]byte(testModuleSchemaOutput), &testSchema)
	// if err != nil {
	// 	t.Fatal(err)
	// }

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
			"experimental": {
				"showReferencesCommandId": "test.id"
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
			"text": %q,
			"uri": "%s/main.tf"
		}
	}`, `variable "test" {
}
output "test" {
	value = var.test
}
`, tmpDir.URI)})

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method: "textDocument/codeLens",
		ReqParams: fmt.Sprintf(`{
			"textDocument": {
				"uri": "%s/main.tf"
			}
		}`, TempDir(t).URI),
	}, `{
				"jsonrpc": "2.0",
				"id": 3,
				"result": [
					{
						"range": {
							"start": {
								"line": 0,
								"character": 0
							},
							"end": {
								"line": 1,
								"character": 1
							}
						},
						"command": {
							"title": "1 reference",
							"command": "test.id",
							"arguments": [
								{
									"line": 0,
									"character": 7
								},
								{
									"includeDeclaration": false
								}
							]
						}
					}
				]
	}`)
}

func TestCodeLens_referenceCount_crossModule(t *testing.T) {
	rootModPath, err := filepath.Abs(filepath.Join("testdata", "single-submodule"))
	if err != nil {
		t.Fatal(err)
	}

	submodPath := filepath.Join(rootModPath, "application")

	rootModUri := document.DirHandleFromPath(rootModPath)
	submodUri := document.DirHandleFromPath(submodPath)

	var testSchema tfjson.ProviderSchemas
	err = json.Unmarshal([]byte(testModuleSchemaOutput), &testSchema)
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
		"capabilities": {
			"experimental": {
				"showReferencesCommandId": "test.id"
			}
		},
		"rootUri": %q,
		"processId": 12345
	}`, rootModUri.URI)})

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
			"text": %q,
			"uri": "%s/main.tf"
		}
	}`, `variable "environment_name" {
  type = string
}

variable "app_prefix" {
  type = string
}

variable "instances" {
  type = number
}
`, submodUri.URI)})

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method: "textDocument/codeLens",
		ReqParams: fmt.Sprintf(`{
			"textDocument": {
				"uri": "%s/main.tf"
			}
		}`, submodUri.URI),
	}, `{
			"jsonrpc": "2.0",
			"id": 3,
			"result": [
				{
					"range": {
						"start": {
							"line": 0,
							"character": 0
						},
						"end": {
							"line": 2,
							"character": 1
						}
					},
					"command": {
						"title": "1 reference",
						"command": "test.id",
						"arguments": [
							{
								"line": 0,
								"character": 13
							},
							{
								"includeDeclaration": false
							}
						]
					}
				},
				{
					"range": {
						"start": {
							"line": 4,
							"character": 0
						},
						"end": {
							"line": 6,
							"character": 1
						}
					},
					"command": {
						"title": "1 reference",
						"command": "test.id",
						"arguments": [
							{
								"line": 4,
								"character": 10
							},
							{
								"includeDeclaration": false
							}
						]
					}
				},
				{
					"range": {
						"start": {
							"line": 8,
							"character": 0
						},
						"end": {
							"line": 10,
							"character": 1
						}
					},
					"command": {
						"title": "1 reference",
						"command": "test.id",
						"arguments": [
							{
								"line": 8,
								"character": 10
							},
							{
								"includeDeclaration": false
							}
						]
					}
				}
			]
	}`)
}
