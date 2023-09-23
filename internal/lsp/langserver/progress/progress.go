// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package progress

import (
	"context"

	"github.com/creachadair/jrpc2"
	gopls "github.com/walteh/retab/gen/gopls/protocol"
	lsctx "github.com/walteh/retab/internal/lsp/context"
)

func Begin(ctx context.Context, title string) error {
	token, ok := lsctx.ProgressToken(ctx)
	if !ok {
		return nil
	}

	return jrpc2.ServerFromContext(ctx).Notify(ctx, "$/progress", gopls.ProgressParams{
		Token: token,
		Value: gopls.WorkDoneProgressBegin{
			Kind:  "begin",
			Title: title,
		},
	})
}

func Report(ctx context.Context, message string) error {
	token, ok := lsctx.ProgressToken(ctx)
	if !ok {
		return nil
	}

	return jrpc2.ServerFromContext(ctx).Notify(ctx, "$/progress", gopls.ProgressParams{
		Token: token,
		Value: gopls.WorkDoneProgressReport{
			Kind:    "report",
			Message: message,
		},
	})
}

func End(ctx context.Context, message string) error {
	token, ok := lsctx.ProgressToken(ctx)
	if !ok {
		return nil
	}

	return jrpc2.ServerFromContext(ctx).Notify(ctx, "$/progress", gopls.ProgressParams{
		Token: token,
		Value: gopls.WorkDoneProgressEnd{
			Kind:    "end",
			Message: message,
		},
	})
}
