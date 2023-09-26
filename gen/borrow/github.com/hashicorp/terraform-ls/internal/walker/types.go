// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package walker

import "github.com/walteh/retab/gen/borrow/github.com/hashicorp/terraform-ls/internal/document"

type DocumentStore interface {
	HasOpenDocuments(dirHandle document.DirHandle) (bool, error)
}
