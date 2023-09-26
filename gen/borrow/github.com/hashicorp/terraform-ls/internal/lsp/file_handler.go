// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lsp

import (
	"github.com/walteh/retab/gen/borrow/github.com/hashicorp/terraform-ls/internal/document"
	lsp "github.com/walteh/retab/gen/borrow/github.com/hashicorp/terraform-ls/internal/protocol"
)

func HandleFromDocumentURI(docUri lsp.DocumentURI) document.Handle {
	return document.HandleFromURI(string(docUri))
}

func DirHandleFromDirURI(dirUri lsp.DocumentURI) document.DirHandle {
	return document.DirHandleFromURI(string(dirUri))
}
