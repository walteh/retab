// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package hooks enables the implementation of hooks for dynamic
// autocompletion. Hooks should be added to this package and
// registered via AppendCompletionHooks in completion_hooks.go.
package hooks

import (
	"log"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/walteh/retab/internal/lsp/registry"
	"github.com/walteh/retab/internal/lsp/state"
)

type Hooks struct {
	ModStore       *state.ModuleStore
	RegistryClient registry.Client
	AlgoliaClient  *search.Client
	Logger         *log.Logger
}
