// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package filesystem

import (
	"github.com/spf13/afero"
)

func (me *Filesystem) Ref() afero.Fs {
	return me.osFs
}
