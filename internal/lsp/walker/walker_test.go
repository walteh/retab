// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package walker

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
	lsctx "github.com/walteh/retab/internal/lsp/context"
	"github.com/walteh/retab/internal/lsp/document"
	"github.com/walteh/retab/internal/lsp/job"
	"github.com/walteh/retab/internal/lsp/state"
)

func TestWalker_basic(t *testing.T) {
	ss, err := state.NewStateStore()
	if err != nil {
		t.Fatal(err)
	}

	// fs := filesystem.NewFilesystem(ss.DocumentStore)
	pa := state.NewPathAwaiter(ss.WalkerPaths, false)

	walkFunc := func(ctx context.Context, modHandle document.DirHandle) (job.IDs, error) {
		return job.IDs{}, nil
	}

	w := NewWalker(afero.NewMemMapFs(), pa, walkFunc)
	w.Collector = NewWalkerCollector()
	w.SetLogger(testLogger())

	root, err := filepath.Abs(filepath.Join("testdata", "uninitialized-root"))
	if err != nil {
		t.Fatal(err)
	}
	dir := document.DirHandleFromPath(root)

	ctx := context.Background()
	err = ss.WalkerPaths.EnqueueDir(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx = lsctx.WithRPCContext(ctx, lsctx.RPCContextData{})
	err = w.StartWalking(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = ss.WalkerPaths.WaitForDirs(ctx, []document.DirHandle{dir})
	if err != nil {
		t.Fatal(err)
	}
	err = ss.JobStore.WaitForJobs(ctx, w.Collector.JobIds()...)
	if err != nil {
		t.Fatal(err)
	}
	err = w.Collector.ErrorOrNil()
	if err != nil {
		t.Fatal(err)
	}
}

func validTfMockCalls(repeatability int) []*mock.Call {
	return []*mock.Call{
		{
			Method: "Version",
			// Repeatability: repeatability,
			Arguments: []interface{}{
				mock.AnythingOfType("*context.valueCtx"),
			},
			ReturnArguments: []interface{}{
				version.Must(version.NewVersion("0.12.0")),
				nil,
				nil,
			},
		},
		{
			Method: "GetExecPath",
			// Repeatability: repeatability,
			ReturnArguments: []interface{}{
				"",
			},
		},
		{
			Method: "ProviderSchemas",
			// Repeatability: repeatability,
			Arguments: []interface{}{
				mock.AnythingOfType("*context.valueCtx"),
			},
			ReturnArguments: []interface{}{
				testProviderSchema,
				nil,
			},
		},
	}
}

var testProviderSchema = &tfjson.ProviderSchemas{
	FormatVersion: "0.1",
	Schemas: map[string]*tfjson.ProviderSchema{
		"test": {
			ConfigSchema: &tfjson.Schema{},
		},
	},
}

func testLogger() *log.Logger {
	if testing.Verbose() {
		return log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	}

	return log.New(ioutil.Discard, "", 0)
}
