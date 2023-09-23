// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifier

import (
	"context"
	"io"
	"log"
)

type moduleCtxKey struct{}
type moduleIsOpenCtxKey struct{}

type Notifier struct {
	hooks  []Hook
	logger *log.Logger
}

type Hook func(ctx context.Context) error

func NewNotifier(hooks []Hook) *Notifier {
	return &Notifier{
		hooks:  hooks,
		logger: defaultLogger,
	}
}

func (n *Notifier) SetLogger(logger *log.Logger) {
	n.logger = logger
}

func (n *Notifier) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				n.logger.Printf("stopping notifier: %s", ctx.Err())
				return
			default:
			}

			err := n.notify(ctx)
			if err != nil {
				n.logger.Printf("failed to notify a change batch: %s", err)
			}
		}
	}()
}

func (n *Notifier) notify(ctx context.Context) error {
	for i, h := range n.hooks {
		err := h(ctx)
		if err != nil {
			n.logger.Printf("hook error (%d): %s", i, err)
			continue
		}
	}

	return nil
}

var defaultLogger = log.New(io.Discard, "", 0)
