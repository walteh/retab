// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package document

import (
	"io/fs"
)

func (fi Document) Name() string {
	return fi.Filename
}

func (fi Document) Size() int64 {
	return int64(len(fi.Text))
}

func (fi Document) Mode() fs.FileMode {
	return 0o755
}

func (fi Document) IsDir() bool {
	return false
}

func (fi Document) Sys() interface{} {
	return nil
}
