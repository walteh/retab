// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	gopls "github.com/walteh/retab/gen/gopls/protocol"
)

func (svc *service) TextDocumentDidSave(ctx context.Context, params gopls.DidSaveTextDocumentParams) error {

	return nil
}
