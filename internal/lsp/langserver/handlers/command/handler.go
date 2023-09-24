// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"log"

	"github.com/walteh/retab/internal/lsp/state"
)

type CmdHandler struct {
	StateStore *state.StateStore
	Logger     *log.Logger
}
