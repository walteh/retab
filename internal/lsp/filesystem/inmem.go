// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package filesystem

import (
	"io/fs"
	"time"
)

type inMemFileInfo struct {
	name    string
	size    int
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (fi inMemFileInfo) Name() string {
	return fi.name
}

func (fi inMemFileInfo) Size() int64 {
	return int64(fi.size)
}

func (fi inMemFileInfo) Mode() fs.FileMode {
	return fi.mode
}

func (fi inMemFileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi inMemFileInfo) IsDir() bool {
	return fi.isDir
}

func (fi inMemFileInfo) Sys() interface{} {
	return nil
}
