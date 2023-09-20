// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package indexer

import (
	"context"

	"github.com/walteh/retab/internal/lsp/document"
	"github.com/walteh/retab/internal/lsp/job"
)

func (idx *Indexer) DocumentChanged(ctx context.Context, modHandle document.DirHandle) (job.IDs, error) {
	ids := make(job.IDs, 0)

	parseId, err := idx.jobStore.EnqueueJob(ctx, job.Job{
		Dir: modHandle,
		Func: func(ctx context.Context) error {
			return nil
		},
		Type:        " op.OpTypeParseModuleConfiguration.String()",
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
		Type:        "op.OpTypeDecodeVarsReferences.String()",
		DependsOn:   job.IDs{parseVarsId},
		IgnoreState: true,
	})
	if err != nil {
		return ids, err
	}
	ids = append(ids, varsRefsId)

	return ids, nil
}

func (idx *Indexer) decodeModule(ctx context.Context, modHandle document.DirHandle, dependsOn job.IDs, ignoreState bool) (job.IDs, error) {
	ids := make(job.IDs, 0)

	metaId, err := idx.jobStore.EnqueueJob(ctx, job.Job{
		Dir: modHandle,
		Func: func(ctx context.Context) error {
			return nil
		},
		Type:        "op.OpTypeLoadModuleMetadata.String()",
		DependsOn:   dependsOn,
		IgnoreState: ignoreState,
		Defer: func(ctx context.Context, jobErr error) (jobIds job.IDs, err error) {
			if jobErr != nil {
				err = jobErr
				return
			}

			eSchemaId, err := idx.jobStore.EnqueueJob(ctx, job.Job{
				Dir: modHandle,
				Func: func(ctx context.Context) error {
					return nil
				},
				Type:        "op.OpTypePreloadEmbeddedSchema.String()",
				IgnoreState: ignoreState,
			})
			if err != nil {
				return
			}
			jobIds = append(jobIds, eSchemaId)

			refOriginsId, err := idx.jobStore.EnqueueJob(ctx, job.Job{
				Dir: modHandle,
				Func: func(ctx context.Context) error {
					return nil
				},
				Type:        "op.OpTypeDecodeReferenceOrigins.String()",
				DependsOn:   job.IDs{eSchemaId},
				IgnoreState: ignoreState,
			})
			jobIds = append(jobIds, refOriginsId)
			return
		},
	})
	if err != nil {
		return ids, err
	}
	ids = append(ids, metaId)

	refTargetsId, err := idx.jobStore.EnqueueJob(ctx, job.Job{
		Dir: modHandle,
		Func: func(ctx context.Context) error {
			return nil
		},
		Type:        "op.OpTypeDecodeReferenceTargets.String()",
		DependsOn:   job.IDs{metaId},
		IgnoreState: ignoreState,
	})
	if err != nil {
		return ids, err
	}
	ids = append(ids, refTargetsId)

	// This job may make an HTTP request, and we schedule it in
	// the low-priority queue, so we don't want to wait for it.
	_, err = idx.jobStore.EnqueueJob(ctx, job.Job{
		Dir: modHandle,
		Func: func(ctx context.Context) error {
			return nil
		},
		Priority: job.LowPriority,
		Type:     "op.OpTypeGetModuleDataFromRegistry.String()",
	})
	if err != nil {
		return ids, err
	}

	return ids, nil
}
