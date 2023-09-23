// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"path/filepath"

	"github.com/creachadair/jrpc2"
	rpch "github.com/creachadair/jrpc2/handler"
	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/spf13/afero"
	"github.com/walteh/retab/gen/gopls"
	lsctx "github.com/walteh/retab/internal/lsp/context"
	idecoder "github.com/walteh/retab/internal/lsp/decoder"
	"github.com/walteh/retab/internal/lsp/filesystem"
	"github.com/walteh/retab/internal/lsp/langserver/session"
	"github.com/walteh/retab/internal/lsp/lsp"
	"github.com/walteh/retab/internal/lsp/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type service struct {
	logger         *log.Logger
	srvCtx         context.Context
	sessCtx        context.Context
	stopSession    context.CancelFunc
	fs             *filesystem.Filesystem
	telemetry      telemetry.Sender
	decoder        *decoder.Decoder
	server         session.Server
	singleFileMode bool
}

var discardLogs = log.New(io.Discard, "", 0)

func NewSession(ctx context.Context, fls afero.Fs) session.Session {
	sessCtx, stopSession := context.WithCancel(ctx)

	svc := &service{
		logger:      discardLogs,
		srvCtx:      ctx,
		sessCtx:     sessCtx,
		stopSession: stopSession,
		telemetry:   &telemetry.NoopSender{},
	}

	svc.fs = filesystem.NewFilesystem(fls)
	svc.fs.SetLogger(svc.logger)
	svc.decoder = decoder.NewDecoder(svc.fs)
	decoderContext := idecoder.DecoderContext(ctx)
	svc.AppendCompletionHooks(decoderContext)
	svc.decoder.SetContext(decoderContext)

	return svc
}

func (svc *service) SetLogger(logger *log.Logger) {
	svc.logger = logger
}

// Assigner builds out the jrpc2.Map according to the LSP protocol
// and passes related dependencies to handlers via context
func (svc *service) Assigner() (jrpc2.Assigner, error) {
	svc.logger.Println("Preparing new session ...")

	sess := session.NewSession(svc.stopSession)

	err := sess.Prepare()
	if err != nil {
		return nil, fmt.Errorf("Unable to prepare session: %w", err)
	}

	svc.telemetry = &telemetry.NoopSender{Logger: svc.logger}

	cc := &gopls.ClientCapabilities{}

	rootDir := ""
	commandPrefix := ""
	clientName := ""
	// var expFeatures settings.ExperimentalFeatures

	m := map[string]rpch.Func{
		"initialize": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.Initialize(req)
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)
			ctx = lsctx.WithRootDirectory(ctx, &rootDir)
			ctx = lsctx.WithCommandPrefix(ctx, &commandPrefix)
			ctx = lsp.ContextWithClientName(ctx, &clientName)
			// ctx = lsctx.WithExperimentalFeatures(ctx, &expFeatures)

			version, ok := lsctx.LanguageServerVersion(svc.srvCtx)
			if ok {
				ctx = lsctx.WithLanguageServerVersion(ctx, version)
			}

			return handle(ctx, req, svc.Initialize)
		},
		"initialized": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.ConfirmInitialization(req)
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)

			return handle(ctx, req, svc.Initialized)
		},
		"textDocument/didChange": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}
			return handle(ctx, req, svc.TextDocumentDidChange)
		},
		"textDocument/didOpen": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}
			return handle(ctx, req, svc.TextDocumentDidOpen)
		},
		"textDocument/didClose": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}
			return handle(ctx, req, svc.TextDocumentDidClose)
		},
		"textDocument/documentSymbol": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)

			return handle(ctx, req, svc.TextDocumentSymbol)
		},
		"textDocument/documentLink": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)
			ctx = lsp.ContextWithClientName(ctx, &clientName)

			return handle(ctx, req, svc.TextDocumentLink)
		},
		"textDocument/declaration": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)

			return handle(ctx, req, svc.GoToDeclaration)
		},
		"textDocument/definition": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)

			return handle(ctx, req, svc.GoToDefinition)
		},
		"textDocument/completion": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)
			// ctx = lsctx.WithExperimentalFeatures(ctx, &expFeatures)

			return handle(ctx, req, svc.TextDocumentComplete)
		},
		"completionItem/resolve": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)
			// ctx = lsctx.WithExperimentalFeatures(ctx, &expFeatures)

			return handle(ctx, req, svc.CompletionItemResolve)
		},
		"textDocument/hover": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)
			ctx = lsp.ContextWithClientName(ctx, &clientName)

			return handle(ctx, req, svc.TextDocumentHover)
		},
		"textDocument/codeAction": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)

			return handle(ctx, req, svc.TextDocumentCodeAction)
		},
		"textDocument/codeLens": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)

			return handle(ctx, req, svc.TextDocumentCodeLens)
		},
		"textDocument/formatting": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			return handle(ctx, req, svc.TextDocumentFormatting)
		},
		"textDocument/signatureHelp": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)

			return handle(ctx, req, svc.SignatureHelp)
		},
		"textDocument/semanticTokens/full": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)

			return handle(ctx, req, svc.TextDocumentSemanticTokensFull)
		},
		"textDocument/didSave": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			// ctx = lsctx.WithDiagnosticsNotifier(ctx, svc.diagsNotifier)
			// ctx = lsctx.WithExperimentalFeatures(ctx, &expFeatures)

			return handle(ctx, req, svc.TextDocumentDidSave)
		},
		"workspace/didChangeWorkspaceFolders": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			return handle(ctx, req, svc.DidChangeWorkspaceFolders)
		},
		"workspace/didChangeWatchedFiles": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			return handle(ctx, req, svc.DidChangeWatchedFiles)
		},
		"textDocument/references": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			return handle(ctx, req, svc.References)
		},
		"workspace/executeCommand": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsctx.WithCommandPrefix(ctx, &commandPrefix)
			ctx = lsctx.WithRootDirectory(ctx, &rootDir)
			// ctx = lsctx.WithDiagnosticsNotifier(ctx, svc.diagsNotifier)
			ctx = lsp.ContextWithClientName(ctx, &clientName)

			return handle(ctx, req, svc.WorkspaceExecuteCommand)
		},
		"workspace/symbol": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)

			return handle(ctx, req, svc.WorkspaceSymbol)
		},
		"shutdown": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.Shutdown(req)
			if err != nil {
				return nil, err
			}
			svc.shutdown()
			return handle(ctx, req, Shutdown)
		},
		"exit": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.Exit()
			if err != nil {
				return nil, err
			}

			svc.stopSession()

			return nil, nil
		},
		"$/cancelRequest": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := sess.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			return handle(ctx, req, CancelRequest)
		},
	}

	return convertMap(m), nil
}

func (svc *service) setupTelemetry(version int, notifier session.ClientNotifier) error {
	t, err := telemetry.NewSender(version, notifier)
	if err != nil {
		return err
	}

	svc.telemetry = t
	return nil
}

func (svc *service) Finish(_ jrpc2.Assigner, status jrpc2.ServerStatus) {
	if status.Closed || status.Err != nil {
		svc.logger.Printf("session stopped unexpectedly (err: %v)", status.Err)
	}

	svc.shutdown()
	svc.stopSession()
}

func (svc *service) shutdown() {

}

// convertMap is a helper function allowing us to omit the jrpc2.Func
// signature from the method definitions
func convertMap(m map[string]rpch.Func) rpch.Map {
	hm := make(rpch.Map, len(m))

	for method, fun := range m {
		hm[method] = rpch.New(fun)
	}

	return hm
}

const requestCancelled jrpc2.Code = -32800
const tracerName = "github.com/walteh/retab/internal/lsp/langserver/handlers"

// handle calls a jrpc2.Func compatible function
func handle(ctx context.Context, req *jrpc2.Request, fn interface{}) (interface{}, error) {
	attrs := []attribute.KeyValue{
		{
			Key:   semconv.RPCMethodKey,
			Value: attribute.StringValue(req.Method()),
		},
		{
			Key:   semconv.RPCJsonrpcRequestIDKey,
			Value: attribute.StringValue(req.ID()),
		},
	}

	// We could capture all parameters here but for now we just
	// opportunistically track the most important ones only.
	type t struct {
		URI string `json:"uri,omitempty"`
	}
	type p struct {
		TextDocument t      `json:"textDocument,omitempty"`
		RootURI      string `json:"rootUri,omitempty"`
	}
	params := p{}
	err := req.UnmarshalParams(&params)
	if err != nil {
		return nil, err
	}

	uri := params.TextDocument.URI
	if params.RootURI != "" {
		uri = params.RootURI
	}

	attrs = append(attrs, attribute.KeyValue{
		Key:   attribute.Key("URI"),
		Value: attribute.StringValue(uri),
	})

	tracer := otel.Tracer(tracerName)
	ctx, span := tracer.Start(ctx, "rpc:"+req.Method(),
		trace.WithAttributes(attrs...))
	defer span.End()

	ctx = lsctx.WithRPCContext(ctx, lsctx.RPCContextData{
		Method: req.Method(),
		URI:    uri,
	})

	result, err := rpch.New(fn)(ctx, req)
	if ctx.Err() != nil && errors.Is(ctx.Err(), context.Canceled) {
		err = fmt.Errorf("%w: %s", requestCancelled.Err(), err)
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "request failed")
	} else {
		span.SetStatus(codes.Ok, "ok")
	}

	return result, err
}

func (svc *service) decoderForDocument(ctx context.Context, name string) (*decoder.PathDecoder, error) {
	return svc.decoder.Path(lang.Path{
		Path:       filepath.Dir(name),
		LanguageID: lsp.Retab.String(),
	})
}
