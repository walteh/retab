// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package utm

import (
	"context"

	"github.com/walteh/retab/internal/lsp/lsp"
)

const UtmSource = "terraform-ls"

func UtmMedium(ctx context.Context) string {
	clientName, ok := lsp.ClientName(ctx)
	if ok {
		return clientName
	}

	return ""
}
