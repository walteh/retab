// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lsp

// LanguageID represents the coding language
// of a file
type LanguageID string

const (
	Retab LanguageID = "retab"
)

func (l LanguageID) String() string {
	return string(l)
}
