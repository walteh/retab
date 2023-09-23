// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lsp

import (
	gopls "github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/lsp/document"
)

func HandleFromDocumentURI(docUri gopls.DocumentURI) document.Handle {
	return document.HandleFromURI(string(docUri))
}

func DirHandleFromDirURI(dirUri gopls.DocumentURI) document.DirHandle {
	return document.DirHandleFromURI(string(dirUri))
}
