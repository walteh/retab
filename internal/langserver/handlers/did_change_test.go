// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/walteh/retab/internal/document"
	"github.com/walteh/retab/internal/langserver"
	"github.com/walteh/retab/internal/state"
)

func TestLangServer_didChange_sequenceOfPartialChanges(t *testing.T) {
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
	    "capabilities": {},
	    "rootUri": %q,
	    "processId": 12345
	}`, tmpDir.URI)})

	ls.Notify(t, &langserver.CallRequest{
		Method:    "initialized",
		ReqParams: "{}",
	})

	originalText := `variable "service_host" {
  default = "blah"
}

module "app" {
  source = "./sub"
  service_listeners = [
    {
      hosts    = [var.service_host]
      listener = ""
    }
  ]
}
`
	ls.Call(t, &langserver.CallRequest{
		Method: "textDocument/didOpen",
		ReqParams: fmt.Sprintf(`{
    "textDocument": {
        "languageId": "terraform",
        "version": 0,
        "uri": "%s/main.tf",
        "text": %q
    }
}`, TempDir(t).URI, originalText)})
	waitForAllJobs(t, ss)

	ls.Call(t, &langserver.CallRequest{
		Method: "textDocument/didChange",
		ReqParams: fmt.Sprintf(`{
    "textDocument": {
        "version": 1,
        "uri": "%s/main.tf"
    },
    "contentChanges": [
        {
            "text": "\n",
            "rangeLength": 0,
            "range": {
                "end": {
                    "line": 8,
                    "character": 18
                },
                "start": {
                    "line": 8,
                    "character": 18
                }
            }
        },
        {
            "text": "      ",
            "rangeLength": 0,
            "range": {
                "end": {
                    "line": 9,
                    "character": 0
                },
                "start": {
                    "line": 9,
                    "character": 0
                }
            }
        }
    ]
}`, TempDir(t).URI)})
	ls.Call(t, &langserver.CallRequest{
		Method: "textDocument/didChange",
		ReqParams: fmt.Sprintf(`{
    "textDocument": {
        "version": 2,
        "uri": "%s/main.tf"
    },
    "contentChanges": [
        {
            "text": "  ",
            "rangeLength": 0,
            "range": {
                "end": {
                    "line": 9,
                    "character": 6
                },
                "start": {
                    "line": 9,
                    "character": 6
                }
            }
        }
    ]
}`, TempDir(t).URI)})

	path := filepath.Join(TempDir(t).Path(), "main.tf")
	dh := document.HandleFromPath(path)
	doc, err := ss.DocumentStore.GetDocument(dh)
	if err != nil {
		t.Fatal(err)
	}

	expectedText := `variable "service_host" {
  default = "blah"
}

module "app" {
  source = "./sub"
  service_listeners = [
    {
      hosts    = [
        var.service_host]
      listener = ""
    }
  ]
}
`

	if diff := cmp.Diff(expectedText, string(doc.Text)); diff != "" {
		t.Fatalf("unexpected text: %s", diff)
	}
}
