// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/walteh/retab/internal/lsp/lsp"
)

type languageIdCtxKey struct{}

func WithLanguageId(ctx context.Context, langId lsp.LanguageID) context.Context {
	return context.WithValue(ctx, languageIdCtxKey{}, langId)
}

func LanguageId(ctx context.Context) (lsp.LanguageID, bool) {
	id, ok := ctx.Value(languageIdCtxKey{}).(lsp.LanguageID)
	if !ok {
		return "", false
	}
	return id, true
}
