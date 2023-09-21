// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lsp

import (
	"context"
	"errors"

	"github.com/walteh/retab/gen/gopls"
)

type clientCapsCtxKey struct{}

func SetClientCapabilities(ctx context.Context, caps *gopls.ClientCapabilities) error {
	cc, ok := ctx.Value(clientCapsCtxKey{}).(*gopls.ClientCapabilities)
	if !ok {
		return errors.New("client capabilities not found")
	}

	*cc = *caps
	return nil
}

func WithClientCapabilities(ctx context.Context, caps *gopls.ClientCapabilities) context.Context {
	return context.WithValue(ctx, clientCapsCtxKey{}, caps)
}

func ClientCapabilities(ctx context.Context) (gopls.ClientCapabilities, error) {
	caps, ok := ctx.Value(clientCapsCtxKey{}).(*gopls.ClientCapabilities)
	if !ok {
		return gopls.ClientCapabilities{}, errors.New("client capabilities not found")
	}

	return *caps, nil
}
