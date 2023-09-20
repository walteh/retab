// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package indexer

import (
	"io"
	"log"

	"github.com/walteh/retab/internal/lsp/job"
)

type Indexer struct {
	logger   *log.Logger
	fs       ReadOnlyFS
	jobStore job.JobStore
}

func NewIndexer(fs ReadOnlyFS, jobStore job.JobStore) *Indexer {

	discardLogger := log.New(io.Discard, "", 0)

	return &Indexer{
		fs:       fs,
		jobStore: jobStore,
		logger:   discardLogger,
	}
}

func (idx *Indexer) SetLogger(logger *log.Logger) {
	idx.logger = logger
}

type Collector interface {
	CollectJobId(jobId job.ID)
}
