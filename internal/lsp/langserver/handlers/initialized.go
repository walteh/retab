// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	lsp "github.com/walteh/retab/gen/gopls"
	ilsp "github.com/walteh/retab/internal/lsp/lsp"
)

func (svc *service) Initialized(ctx context.Context, params lsp.InitializedParams) error {
	caps, err := ilsp.ClientCapabilities(ctx)
	if err != nil {
		return err
	}

	return svc.setupWatchedFiles(ctx, caps.Workspace.DidChangeWatchedFiles)
}

func (svc *service) setupWatchedFiles(ctx context.Context, caps lsp.DidChangeWatchedFilesClientCapabilities) error {
	if !caps.DynamicRegistration {
		svc.logger.Printf("Client doesn't support dynamic watched files registration, " +
			"provider and module changes may not be reflected at runtime")
		return nil
	}

	// id, err := uuid.GenerateUUID()
	// if err != nil {
	// 	return err
	// }

	// watchers := make([]lsp.FileSystemWatcher, len(watchPatterns))
	// for i, wp := range watchPatterns {
	// 	watchers[i] = lsp.FileSystemWatcher{
	// 		GlobPattern: wp.Pattern,
	// 		Kind:        kindFromEventType(wp.EventType),
	// 	}
	// }

	// srv := jrpc2.ServerFromContext(ctx)
	// _, err = srv.Callback(ctx, "client/registerCapability", lsp.RegistrationParams{
	// 	Registrations: []lsp.Registration{
	// 		{
	// 			ID:     id,
	// 			Method: "workspace/didChangeWatchedFiles",
	// 			RegisterOptions: lsp.DidChangeWatchedFilesRegistrationOptions{
	// 				Watchers: watchers,
	// 			},
	// 		},
	// 	},
	// })
	// if err != nil {
	// 	svc.logger.Printf("failed to register watched files: %s", err)
	// }
	return nil
}
