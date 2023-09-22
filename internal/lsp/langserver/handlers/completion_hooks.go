// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"github.com/hashicorp/hcl-lang/decoder"
)

func (s *service) AppendCompletionHooks(decoderContext decoder.DecoderContext) {
	// h := hooks.Hooks{
	// 	Logger: s.logger,
	// }

	// decoderContext.CompletionHooks["CompleteLocalModuleSources"] = h.LocalModuleSources
	// decoderContext.CompletionHooks["CompleteRegistryModuleSources"] = h.RegistryModuleSources
	// decoderContext.CompletionHooks["CompleteRegistryModuleVersions"] = h.RegistryModuleVersions
}
