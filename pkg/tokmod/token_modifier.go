// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tokmod

import (
	"github.com/hashicorp/hcl-lang/lang"
)

var (
	File = lang.SemanticTokenModifier("retab-file")
)

var SupportedModifiers = []lang.SemanticTokenModifier{
	File,
}
