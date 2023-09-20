// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package indexer

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/walteh/retab/internal/lsp/document"
	"github.com/walteh/retab/internal/lsp/job"
)

func (idx *Indexer) DocumentOpened(ctx context.Context, modHandle document.DirHandle) (job.IDs, error) {

	ids := make(job.IDs, 0)
	var errs *multierror.Error

	parseId, err := idx.jobStore.EnqueueJob(ctx, job.Job{
		Dir: modHandle,
		Func: func(ctx context.Context) error {
			return nil
		},
		Type:        "op.OpTypeParseModuleConfiguration.String()",
		IgnoreState: true,
	})
	if err != nil {
		return ids, err
	}
	ids = append(ids, parseId)

	modIds, err := idx.decodeModule(ctx, modHandle, job.IDs{parseId}, true)
	if err != nil {
		return ids, err
	}
	ids = append(ids, modIds...)

	parseVarsId, err := idx.jobStore.EnqueueJob(ctx, job.Job{
		Dir: modHandle,
		Func: func(ctx context.Context) error {
			return nil
		},
		Type:        "op.OpTypeParseVariables.String()",
		IgnoreState: true,
	})
	if err != nil {
		return ids, err
	}
	ids = append(ids, parseVarsId)

	varsRefsId, err := idx.jobStore.EnqueueJob(ctx, job.Job{
		Dir: modHandle,
		Func: func(ctx context.Context) error {
			return nil
		},
		Type:      "op.OpTypeDecodeVarsReferences.String()",
		DependsOn: job.IDs{parseVarsId},
	})
	if err != nil {
		return ids, err
	}
	ids = append(ids, varsRefsId)

	return ids, errs.ErrorOrNil()
}
