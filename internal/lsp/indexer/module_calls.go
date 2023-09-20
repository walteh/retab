// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package indexer

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/walteh/retab/internal/lsp/document"
	"github.com/walteh/retab/internal/lsp/job"
)

func (idx *Indexer) collectReferences(ctx context.Context, modHandle document.DirHandle, dependsOn job.IDs, ignoreState bool) (job.IDs, error) {
	ids := make(job.IDs, 0)

	var errs *multierror.Error

	id, err := idx.jobStore.EnqueueJob(ctx, job.Job{
		Dir: modHandle,
		Func: func(ctx context.Context) error {
			return nil
		},
		Type:        "decodeReferenceTargets",
		DependsOn:   dependsOn,
		IgnoreState: ignoreState,
	})
	if err != nil {
		errs = multierror.Append(errs, err)
	} else {
		ids = append(ids, id)
	}

	id, err = idx.jobStore.EnqueueJob(ctx, job.Job{
		Dir: modHandle,
		Func: func(ctx context.Context) error {
			return nil
		},
		Type:        "collectReferences",
		DependsOn:   dependsOn,
		IgnoreState: ignoreState,
	})
	if err != nil {
		errs = multierror.Append(errs, err)
	} else {
		ids = append(ids, id)
	}

	return ids, errs.ErrorOrNil()
}
