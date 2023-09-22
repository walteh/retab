// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/walteh/retab/internal/lsp/langserver"
	"github.com/walteh/retab/internal/lsp/state"
)

func TestDocumentLink_withValidData(t *testing.T) {
	tmpDir := TempDir(t)
	err := os.WriteFile(filepath.Join(tmpDir.Path(), "main.tf"), []byte("provider \"test\" {\n\n}\n"), 0o755)
	if err != nil {
		t.Fatal(err)
	}

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
		Method: "textDocument/documentLink",
		ReqParams: fmt.Sprintf(`{
			"textDocument": {
				"uri": "%s/main.tf"
			}
		}`, tmpDir.URI)}, `{
			"jsonrpc": "2.0",
			"id": 3,
			"result": [
				{
					"range": {
						"start": {
							"line": 0,
							"character": 9
						},
						"end": {
							"line": 0,
							"character": 15
						}
					},
					"target": "https://registry.terraform.io/providers/test/test/latest/docs?utm_content=documentLink\u0026utm_source=terraform-ls"
				}
			]
		}`)
}
