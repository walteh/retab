// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/walteh/retab/internal/langserver"
	"github.com/walteh/retab/internal/state"
)

func TestLangServer_DidChangeWatchedFiles_change_file(t *testing.T) {
	tmpDir := TempDir(t)

	InitPluginCache(t, tmpDir.Path())

	originalSrc := `variable "original" {
  default = "foo"
}
`
	err := os.WriteFile(filepath.Join(tmpDir.Path(), "main.tf"), []byte(originalSrc), 0o755)
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
	}`, testHandle.URI)})

	ls.Notify(t, &langserver.CallRequest{
		Method:    "initialized",
		ReqParams: "{}",
	})

	ls.Call(t, &langserver.CallRequest{
		Method: "workspace/didChangeWatchedFiles",
		ReqParams: fmt.Sprintf(`{
    "changes": [
        {
            "uri": "%s/.terraform/modules/modules.json",
            "type": 1
        }
    ]
}`, testHandle.URI)})

}
