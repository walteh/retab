// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	"github.com/walteh/retab/gen/gopls"
	"github.com/walteh/retab/internal/lsp/document"
	"github.com/walteh/retab/internal/lsp/job"
	"github.com/walteh/retab/internal/lsp/uri"
)

func (svc *service) DidChangeWatchedFiles(ctx context.Context, params gopls.DidChangeWatchedFilesParams) error {
	var ids job.IDs

	for _, change := range params.Changes {
		rawURI := string(change.URI)

		_, err := uri.PathFromURI(rawURI)
		if err != nil {
			svc.logger.Printf("error parsing %q: %s", rawURI, err)
			continue
		}

		if change.Type == gopls.Deleted {

		}

		if change.Type == gopls.Changed {

		}

		if change.Type == gopls.Created {
		}
	}

	err := svc.stateStore.JobStore.WaitForJobs(ctx, ids...)
	if err != nil {
		return err
	}

	return nil
}

type parsedModuleHandle struct {
	DirHandle document.DirHandle
	IsDir     bool
}

var ErrorSkip = errSkip{}

type errSkip struct{}

func (es errSkip) Error() string {
	return "skip"
}
