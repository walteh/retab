// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"fmt"
	"testing"

	"github.com/walteh/retab/internal/lsp/langserver"
	"github.com/walteh/retab/internal/lsp/langserver/session"
)

func TestLangServer_didOpenWithoutInitialization(t *testing.T) {
	ls := langserver.NewLangServerMock(t, NewMockSession(nil))
	stop := ls.Start(t)
	defer stop()

	ls.CallAndExpectError(t, &langserver.CallRequest{
		Method: "textDocument/didOpen",
		ReqParams: fmt.Sprintf(`{
		"textDocument": {
			"version": 0,
			"languageId": "retab",
			"text": "provider \"github\" {}",
			"uri": "%s/main.tf"
		}
	}`, TempDir(t).URI)}, session.SessionNotInitialized.Err())
}

// func TestLangServer_didOpenLanguageIdStored(t *testing.T) {
// 	tmpDir := TempDir(t)

// 	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{}))
// 	stop := ls.Start(t)
// 	defer stop()

// 	ls.Call(t, &langserver.CallRequest{
// 		Method: "initialize",
// 		ReqParams: fmt.Sprintf(`{
// 	    "capabilities": {},
// 	    "rootUri": %q,
// 	    "processId": 12345
// 	}`, tmpDir.URI)})

// 	ls.Notify(t, &langserver.CallRequest{
// 		Method:    "initialized",
// 		ReqParams: "{}",
// 	})

// 	originalText := `variable "service_host" {
//   default = "blah"
// }
// `
// 	ls.Call(t, &langserver.CallRequest{
// 		Method: "textDocument/didOpen",
// 		ReqParams: fmt.Sprintf(`{
//     "textDocument": {
//         "languageId": "retab",
//         "version": 0,
//         "uri": "%s/main.tf",
//         "text": %q
//     }
// }`, TempDir(t).URI, originalText)})

// 	path := filepath.Join(TempDir(t).Path(), "main.tf")
// 	filename := string(path)

// 	// if diff := cmp.Diff(doc.LanguageID, string("retab")); diff != "" {
// 	// 	t.Fatalf("unexpected languageID: %s", diff)
// 	// }
// 	fullPath := doc.FullPath()
// 	if diff := cmp.Diff(fullPath, string(path)); diff != "" {
// 		t.Fatalf("unexpected fullPath: %s", diff)
// 	}
// 	version := doc.Version
// 	if diff := cmp.Diff(version, int(0)); diff != "" {
// 		t.Fatalf("unexpected version: %s", diff)
// 	}
// }
