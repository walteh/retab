// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/walteh/retab/internal/lsp/document"
	"github.com/walteh/retab/internal/lsp/langserver"
	"github.com/walteh/retab/internal/lsp/langserver/session"
	"github.com/walteh/retab/internal/lsp/state"
)

func BenchmarkInitializeFolder_basic(b *testing.B) {
	b.Skip()

	modules := []struct {
		name       string
		sourceAddr string
	}{
		{
			name:       "local-single-module-no-provider",
			sourceAddr: mustAbs(b, filepath.Join("testdata", "single-module-no-provider")),
		},
		{
			name:       "local-single-submodule-no-provider",
			sourceAddr: mustAbs(b, filepath.Join("testdata", "single-submodule")),
		},
		{
			name:       "local-single-module-random",
			sourceAddr: mustAbs(b, filepath.Join("testdata", "single-module-random")),
		},
		{
			name:       "local-single-module-aws",
			sourceAddr: mustAbs(b, filepath.Join("testdata", "single-module-aws")),
		},
		// TODO: module version pinning - requires explicit git cloning
		{
			name:       "aws-consul",
			sourceAddr: "github.com/hashicorp/terraform-aws-consul?ref=v0.11.0",
		},
		{
			name:       "aws-eks",
			sourceAddr: "terraform-aws-modules/eks/aws",
		},
		{
			name:       "aws-vpc",
			sourceAddr: "terraform-aws-modules/vpc/aws",
		},
		{
			name:       "google-project",
			sourceAddr: "terraform-google-modules/project-factory/google",
		},
		{
			name:       "google-network",
			sourceAddr: "terraform-google-modules/network/google",
		},
		{
			name:       "google-gke",
			sourceAddr: "terraform-google-modules/kubernetes-engine/google",
		},
		{
			name:       "k8s-metrics-server",
			sourceAddr: "cookielab/metrics-server/kubernetes",
		},
		{
			name:       "k8s-dashboard",
			sourceAddr: "cookielab/dashboard/kubernetes",
		},
	}

	for _, mod := range modules {
		b.Run(mod.name, func(b *testing.B) {
			rootDir := b.TempDir()

			b.Cleanup(func() {
				os.RemoveAll(rootDir)
			})
			b.StopTimer()
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				rootDir := document.DirHandleFromPath(rootDir)
				ss, err := state.NewStateStore()
				if err != nil {
					b.Fatal(err)
				}

				b.StartTimer()
				ls := langserver.NewLangServerMock(b, func(ctx context.Context) session.Session {
					sessCtx, stopSession := context.WithCancel(ctx)
					return &service{
						logger:      discardLogs,
						srvCtx:      ctx,
						sessCtx:     sessCtx,
						stopSession: stopSession,
						stateStore:  ss,
					}
				})
				stop := ls.Start(b)

				ls.Call(b, &langserver.CallRequest{
					Method: "initialize",
					ReqParams: fmt.Sprintf(`{
						"capabilities": {
							"workspace": {
								"workspaceFolders": true
							}
						},
						"rootUri": %q,
						"processId": 12345,
						"workspaceFolders": [
							{
								"uri": %q,
								"name": "root"
							}
						]
					}`, rootDir.URI, rootDir.URI)})

				b.StopTimer()

				stop()
			}
		})
	}
}

func mustAbs(b *testing.B, path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		b.Fatal(err)
	}
	return absPath
}
