// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package session

import (
	"context"

	"github.com/creachadair/jrpc2"
)

type ClientNotifier interface {
	Notify(ctx context.Context, method string, params interface{}) error
}

type ClientCaller interface {
	Callback(ctx context.Context, method string, params interface{}) (*jrpc2.Response, error)
}

type Server interface {
	ClientNotifier
	ClientCaller
}
