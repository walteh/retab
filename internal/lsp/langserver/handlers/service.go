// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/creachadair/jrpc2"
	rpch "github.com/creachadair/jrpc2/handler"
	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/walteh/retab/gen/gopls"
	lsctx "github.com/walteh/retab/internal/lsp/context"
	idecoder "github.com/walteh/retab/internal/lsp/decoder"
	"github.com/walteh/retab/internal/lsp/document"
	"github.com/walteh/retab/internal/lsp/filesystem"
	"github.com/walteh/retab/internal/lsp/indexer"
	"github.com/walteh/retab/internal/lsp/job"
	"github.com/walteh/retab/internal/lsp/langserver/diagnostics"
	"github.com/walteh/retab/internal/lsp/langserver/notifier"
	"github.com/walteh/retab/internal/lsp/langserver/session"
	"github.com/walteh/retab/internal/lsp/lsp"
	"github.com/walteh/retab/internal/lsp/scheduler"
	"github.com/walteh/retab/internal/lsp/settings"
	"github.com/walteh/retab/internal/lsp/state"
	"github.com/walteh/retab/internal/lsp/telemetry"
	"github.com/walteh/retab/internal/lsp/walker"
	"github.com/walteh/retab/internal/protocol"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type service struct {
	logger *log.Logger

	srvCtx context.Context

	sessCtx     context.Context
	stopSession context.CancelFunc

	// TODO: Rename to *scheduler to avoid confusion
	lowPrioIndexer  *scheduler.Scheduler
	highPrioIndexer *scheduler.Scheduler

	closedDirWalker *walker.Walker
	openDirWalker   *walker.Walker

	fs            *filesystem.Filesystem
	telemetry     telemetry.Sender
	decoder       *decoder.Decoder
	stateStore    *state.StateStore
	server        session.Server
	diagsNotifier *diagnostics.Notifier
	notifier      *notifier.Notifier
	indexer       *indexer.Indexer

	walkerCollector    *walker.WalkerCollector
	additionalHandlers map[string]rpch.Func

	singleFileMode bool
}

var discardLogs = log.New(io.Discard, "", 0)

func NewSession(srvCtx context.Context) session.Session {
	sessCtx, stopSession := context.WithCancel(srvCtx)
	return &service{
		logger:      discardLogs,
		srvCtx:      srvCtx,
		sessCtx:     sessCtx,
		stopSession: stopSession,
		telemetry:   &telemetry.NoopSender{},
	}
}

func (svc *service) SetLogger(logger *log.Logger) {
	svc.logger = logger
}

// Assigner builds out the jrpc2.Map according to the LSP protocol
// and passes related dependencies to handlers via context
func (svc *service) Assigner() (jrpc2.Assigner, error) {
	svc.logger.Println("Preparing new session ...")

	session := session.NewSession(svc.stopSession)

	err := session.Prepare()
	if err != nil {
		return nil, fmt.Errorf("Unable to prepare session: %w", err)
	}

	svc.telemetry = &telemetry.NoopSender{Logger: svc.logger}

	cc := &gopls.ClientCapabilities{}

	rootDir := ""
	commandPrefix := ""
	clientName := ""
	var expFeatures settings.ExperimentalFeatures

	m := map[string]rpch.Func{
		"initialize": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.Initialize(req)
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)
			ctx = lsctx.WithRootDirectory(ctx, &rootDir)
			ctx = lsctx.WithCommandPrefix(ctx, &commandPrefix)
			ctx = lsp.ContextWithClientName(ctx, &clientName)
			ctx = lsctx.WithExperimentalFeatures(ctx, &expFeatures)

			version, ok := lsctx.LanguageServerVersion(svc.srvCtx)
			if ok {
				ctx = lsctx.WithLanguageServerVersion(ctx, version)
			}

			return handle(ctx, req, svc.Initialize)
		},
		"initialized": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.ConfirmInitialization(req)
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)

			return handle(ctx, req, svc.Initialized)
		},
		"textDocument/didChange": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}
			return handle(ctx, req, svc.TextDocumentDidChange)
		},
		"textDocument/didOpen": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}
			return handle(ctx, req, svc.TextDocumentDidOpen)
		},
		"textDocument/didClose": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}
			return handle(ctx, req, svc.TextDocumentDidClose)
		},
		"textDocument/documentSymbol": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)

			return handle(ctx, req, svc.TextDocumentSymbol)
		},
		"textDocument/documentLink": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)
			ctx = lsp.ContextWithClientName(ctx, &clientName)

			return handle(ctx, req, svc.TextDocumentLink)
		},
		"textDocument/declaration": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)

			return handle(ctx, req, svc.GoToDeclaration)
		},
		"textDocument/definition": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)

			return handle(ctx, req, svc.GoToDefinition)
		},
		"textDocument/completion": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)
			ctx = lsctx.WithExperimentalFeatures(ctx, &expFeatures)

			return handle(ctx, req, svc.TextDocumentComplete)
		},
		"completionItem/resolve": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)
			ctx = lsctx.WithExperimentalFeatures(ctx, &expFeatures)

			return handle(ctx, req, svc.CompletionItemResolve)
		},
		"textDocument/hover": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)
			ctx = lsp.ContextWithClientName(ctx, &clientName)

			return handle(ctx, req, svc.TextDocumentHover)
		},
		"textDocument/codeAction": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)

			return handle(ctx, req, svc.TextDocumentCodeAction)
		},
		"textDocument/codeLens": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)

			return handle(ctx, req, svc.TextDocumentCodeLens)
		},
		"textDocument/formatting": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			return handle(ctx, req, svc.TextDocumentFormatting)
		},
		"textDocument/signatureHelp": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)

			return handle(ctx, req, svc.SignatureHelp)
		},
		"textDocument/semanticTokens/full": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)

			return handle(ctx, req, svc.TextDocumentSemanticTokensFull)
		},
		"textDocument/didSave": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsctx.WithDiagnosticsNotifier(ctx, svc.diagsNotifier)
			ctx = lsctx.WithExperimentalFeatures(ctx, &expFeatures)

			return handle(ctx, req, svc.TextDocumentDidSave)
		},
		"workspace/didChangeWorkspaceFolders": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			return handle(ctx, req, svc.DidChangeWorkspaceFolders)
		},
		"workspace/didChangeWatchedFiles": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			return handle(ctx, req, svc.DidChangeWatchedFiles)
		},
		"textDocument/references": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			return handle(ctx, req, svc.References)
		},
		"workspace/executeCommand": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsctx.WithCommandPrefix(ctx, &commandPrefix)
			ctx = lsctx.WithRootDirectory(ctx, &rootDir)
			ctx = lsctx.WithDiagnosticsNotifier(ctx, svc.diagsNotifier)
			ctx = lsp.ContextWithClientName(ctx, &clientName)

			return handle(ctx, req, svc.WorkspaceExecuteCommand)
		},
		"workspace/symbol": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			ctx = lsp.WithClientCapabilities(ctx, cc)

			return handle(ctx, req, svc.WorkspaceSymbol)
		},
		"shutdown": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.Shutdown(req)
			if err != nil {
				return nil, err
			}
			svc.shutdown()
			return handle(ctx, req, Shutdown)
		},
		"exit": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.Exit()
			if err != nil {
				return nil, err
			}

			svc.stopSession()

			return nil, nil
		},
		"$/cancelRequest": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			err := session.CheckInitializationIsConfirmed()
			if err != nil {
				return nil, err
			}

			return handle(ctx, req, CancelRequest)
		},
	}

	// For use in tests, e.g. to test request cancellation
	if len(svc.additionalHandlers) > 0 {
		for methodName, handlerFunc := range svc.additionalHandlers {
			m[methodName] = handlerFunc
		}
	}

	return convertMap(m), nil
}

func (svc *service) configureSessionDependencies(ctx context.Context, cfgOpts *settings.Options) error {

	if svc.stateStore == nil {
		store, err := state.NewStateStore()
		if err != nil {
			return err
		}
		svc.stateStore = store
	}

	svc.stateStore.SetLogger(svc.logger)

	moduleHooks := []notifier.Hook{
		updateDiagnostics(svc.diagsNotifier),
		sendModuleTelemetry(svc.stateStore, svc.telemetry),
	}

	svc.lowPrioIndexer = scheduler.NewScheduler(svc.stateStore.JobStore, 1, job.LowPriority)
	svc.lowPrioIndexer.SetLogger(svc.logger)
	svc.lowPrioIndexer.Start(svc.sessCtx)
	svc.logger.Printf("started low priority scheduler")

	svc.highPrioIndexer = scheduler.NewScheduler(svc.stateStore.JobStore, 1, job.HighPriority)
	svc.highPrioIndexer.SetLogger(svc.logger)
	svc.highPrioIndexer.Start(svc.sessCtx)
	svc.logger.Printf("started high priority scheduler")

	cc, err := lsp.ClientCapabilities(ctx)
	if err == nil {
		if _, ok := protocol.ExperimentalClientCapabilities(cc.Experimental).ShowReferencesCommandId(); ok {
			moduleHooks = append(moduleHooks, refreshCodeLens(svc.server))
		}

		if commandId, ok := protocol.ExperimentalClientCapabilities(cc.Experimental).RefreshModuleProvidersCommandId(); ok {
			moduleHooks = append(moduleHooks, callRefreshClientCommand(svc.server, commandId))
		}

		if commandId, ok := protocol.ExperimentalClientCapabilities(cc.Experimental).RefreshModuleCallsCommandId(); ok {
			moduleHooks = append(moduleHooks, callRefreshClientCommand(svc.server, commandId))
		}

		if commandId, ok := protocol.ExperimentalClientCapabilities(cc.Experimental).RefreshTerraformVersionCommandId(); ok {
			moduleHooks = append(moduleHooks, callRefreshClientCommand(svc.server, commandId))
		}

		if cc.Workspace.SemanticTokens != nil && cc.Workspace.SemanticTokens.RefreshSupport {
			moduleHooks = append(moduleHooks, refreshSemanticTokens(svc.server))
		}
	}

	svc.notifier = notifier.NewNotifier(moduleHooks)
	svc.notifier.SetLogger(svc.logger)
	svc.notifier.Start(svc.sessCtx)

	svc.fs = filesystem.NewFilesystem(svc.stateStore.DocumentStore)
	svc.fs.SetLogger(svc.logger)

	svc.indexer = indexer.NewIndexer(svc.fs.Ref(), svc.stateStore.JobStore)
	svc.indexer.SetLogger(svc.logger)

	svc.decoder = decoder.NewDecoder(svc.fs)
	decoderContext := idecoder.DecoderContext(ctx)
	svc.AppendCompletionHooks(decoderContext)
	svc.decoder.SetContext(decoderContext)

	closedPa := state.NewPathAwaiter(svc.stateStore.WalkerPaths, false)
	svc.closedDirWalker = walker.NewWalker(svc.fs.Ref(), closedPa, svc.indexer.WalkedModule)
	svc.closedDirWalker.Collector = svc.walkerCollector
	svc.closedDirWalker.SetLogger(svc.logger)

	opendPa := state.NewPathAwaiter(svc.stateStore.WalkerPaths, true)
	svc.openDirWalker = walker.NewWalker(svc.fs.Ref(), opendPa, svc.indexer.WalkedModule)
	svc.closedDirWalker.Collector = svc.walkerCollector
	svc.openDirWalker.SetLogger(svc.logger)

	return nil
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
	if svc.closedDirWalker != nil {
		svc.logger.Printf("stopping closedDirWalker for session ...")
		svc.closedDirWalker.Stop()
		svc.logger.Printf("closedDirWalker stopped")
	}
	if svc.openDirWalker != nil {
		svc.logger.Printf("stopping openDirWalker for session ...")
		svc.openDirWalker.Stop()
		svc.logger.Printf("openDirWalker stopped")
	}

	if svc.lowPrioIndexer != nil {
		svc.lowPrioIndexer.Stop()
	}
	if svc.highPrioIndexer != nil {
		svc.highPrioIndexer.Stop()
	}
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

func (svc *service) decoderForDocument(ctx context.Context, doc *document.Document) (*decoder.PathDecoder, error) {
	return svc.decoder.Path(lang.Path{
		Path:       doc.Dir.Path(),
		LanguageID: doc.LanguageID,
	})
}
