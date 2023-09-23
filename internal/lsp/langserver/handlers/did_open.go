// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"bytes"
	"context"
	"fmt"

	"github.com/creachadair/jrpc2"
	"github.com/spf13/afero"
	gopls "github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/lsp/uri"
)

func (svc *service) TextDocumentDidOpen(ctx context.Context, params gopls.DidOpenTextDocumentParams) error {
	docURI := string(params.TextDocument.URI)

	// URIs are always checked during initialize request, but
	// we still allow single-file mode, therefore invalid URIs
	// can still land here, so we check for those.
	if !uri.IsURIValid(docURI) {
		jrpc2.ServerFromContext(ctx).Notify(ctx, "window/showMessage", &gopls.ShowMessageParams{
			Type: gopls.Warning,
			Message: fmt.Sprintf("Ignoring workspace folder (unsupport or invalid URI) %s."+
				" This is most likely bug, please report it.", docURI),
		})
		return fmt.Errorf("invalid URI: %s", docURI)
	}

	// dh := document.HandleFromURI(docURI)

	// err := svc.stateStore.DocumentStore.OpenDocument(dh, params.TextDocument.LanguageID,
	// 	int(params.TextDocument.Version), []byte(params.TextDocument.Text))
	// if err != nil {
	// 	return err
	// }

	// ctx = lsctx.WithLanguageId(ctx, params.TextDocument.LanguageID)

	return afero.WriteReader(svc.fs, docURI, bytes.NewReader([]byte(params.TextDocument.Text)))

}
