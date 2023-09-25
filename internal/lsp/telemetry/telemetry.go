// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package telemetry

import (
	"context"
	"fmt"

	"github.com/walteh/retab/internal/lsp/protocol"
)

type Telemetry struct {
	version  int
	notifier Notifier
}

type Notifier interface {
	Notify(ctx context.Context, method string, params interface{}) error
}

func NewSender(version int, notifier Notifier) (*Telemetry, error) {
	if version != protocol.TelemetryFormatVersion {
		return nil, fmt.Errorf("unsupported telemetry format version: %d", version)
	}

	return &Telemetry{
		version:  version,
		notifier: notifier,
	}, nil
}

func (t *Telemetry) SendEvent(ctx context.Context, name string, properties map[string]interface{}) {
	t.notifier.Notify(ctx, "telemetry/event", protocol.TelemetryEvent{
		Version:    t.version,
		Name:       name,
		Properties: properties,
	})
}
