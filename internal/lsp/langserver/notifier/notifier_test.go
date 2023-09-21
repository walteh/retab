// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifier

import (
	"context"
	"io"
	"log"
	"sync"
	"testing"
)

func TestNotifier(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	notifier := NewNotifier([]Hook{})
	notifier.SetLogger(testLogger())

	notifier.Start(ctx)

	wg.Wait()
}

func testLogger() *log.Logger {
	if testing.Verbose() {
		return log.Default()
	}
	return log.New(io.Discard, "", 0)
}
