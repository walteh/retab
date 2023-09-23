// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package protocol

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/walteh/retab/gen/gopls/protocol"
)

type CompletionItemWithResolveHook struct {
	protocol.CompletionItem

	ResolveHook *lang.ResolveHook `json:"data,omitempty"`
}
