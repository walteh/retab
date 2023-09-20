// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package indexer

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/walteh/retab/internal/lsp/document"
	"github.com/walteh/retab/internal/lsp/job"
)

func (idx *Indexer) WalkedModule(ctx context.Context, modHandle document.DirHandle) (job.IDs, error) {
	ids := make(job.IDs, 0)
	var errs *multierror.Error

	refCollectionDeps := make(job.IDs, 0)
	providerVersionDeps := make(job.IDs, 0)

	parseId, err := idx.jobStore.EnqueueJob(ctx, job.Job{
		Dir: modHandle,
		Func: func(ctx context.Context) error {
			return nil
		},
		Type: "parseModule",
	})
	if err != nil {
		errs = multierror.Append(errs, err)
	} else {
		ids = append(ids, parseId)
		refCollectionDeps = append(refCollectionDeps, parseId)
		providerVersionDeps = append(providerVersionDeps, parseId)
	}

	var metaId job.ID
	if parseId != "" {
		metaId, err = idx.jobStore.EnqueueJob(ctx, job.Job{
			Dir:  modHandle,
			Type: "op.OpTypeLoadModuleMetadata.String()",
			Func: func(ctx context.Context) error {
				return nil
			},
			DependsOn: job.IDs{parseId},
		})
		if err != nil {
			return ids, err
		} else {
			ids = append(ids, metaId)
			refCollectionDeps = append(refCollectionDeps, metaId)
			providerVersionDeps = append(providerVersionDeps, metaId)
		}
	}

	parseVarsId, err := idx.jobStore.EnqueueJob(ctx, job.Job{
		Dir: modHandle,
		Func: func(ctx context.Context) error {
			return nil
		},
		Type: "op.OpTypeParseVariables.String()",
	})
	if err != nil {
		errs = multierror.Append(errs, err)
	} else {
		ids = append(ids, parseVarsId)
	}

	if parseVarsId != "" {
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
		} else {
			ids = append(ids, varsRefsId)
			refCollectionDeps = append(refCollectionDeps, varsRefsId)
		}
	}

	eSchemaId, err := idx.jobStore.EnqueueJob(ctx, job.Job{
		Dir: modHandle,
		Func: func(ctx context.Context) error {
			return nil
		},
		// This could theoretically also depend on ObtainSchema to avoid
		// attempt to preload the same schema twice but we avoid that dependency
		// as obtaining schema via CLI often takes a long time (multiple
		// seconds) and this would then defeat the main benefit
		// of preloaded schemas which can be loaded in miliseconds.
		DependsOn: providerVersionDeps,
		Type:      "op.OpTypePreloadEmbeddedSchema.String()",
	})
	if err != nil {
		return ids, err
	}
	ids = append(ids, eSchemaId)

	if parseId != "" {
		rIds, err := idx.collectReferences(ctx, modHandle, refCollectionDeps, false)
		if err != nil {
			errs = multierror.Append(errs, err)
		} else {
			ids = append(ids, rIds...)
		}
	}

	return ids, errs.ErrorOrNil()
}
