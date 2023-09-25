package server

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/internal/adapter"
	"github.com/walteh/retab/internal/lsp/lsp"
)

var _ protocol.Server = (*Server)(nil)

// CodeAction implements protocol.Server.
func (*Server) CodeAction(context.Context, *protocol.CodeActionParams) ([]protocol.CodeAction, error) {
	panic("unimplemented")
}

// CodeLens implements protocol.Server.
func (me *Server) CodeLens(ctx context.Context, params *protocol.CodeLensParams) ([]protocol.CodeLens, error) {
	list := make([]protocol.CodeLens, 0)

	filename := string(params.TextDocument.URI)

	path := lang.Path{
		Path:       filename,
		LanguageID: lsp.Retab.String(),
	}

	decd := decoder.NewDecoder(adapter.NewCacheSessionDecoder(me.session))

	lenses, err := decd.CodeLensesForFile(ctx, path, filename)
	if err != nil {
		return nil, err
	}

	for _, lens := range lenses {
		cmd, err := lsp.Command(lens.Command)
		if err != nil {
			fmt.Printf("skipping code lens %#v: %s", lens.Command, err)
			continue
		}

		list = append(list, protocol.CodeLens{
			Range:   lsp.HCLRangeToLSP(lens.Range),
			Command: &cmd,
		})
	}

	return list, nil
}

// ColorPresentation implements protocol.Server.
func (*Server) ColorPresentation(context.Context, *protocol.ColorPresentationParams) ([]protocol.ColorPresentation, error) {
	panic("unimplemented")
}

// Completion implements protocol.Server.
func (*Server) Completion(context.Context, *protocol.CompletionParams) (*protocol.CompletionList, error) {
	panic("unimplemented")
}

// Declaration implements protocol.Server.
func (*Server) Declaration(context.Context, *protocol.DeclarationParams) (*protocol.Or_textDocument_declaration, error) {
	panic("unimplemented")
}

// Definition implements protocol.Server.
func (*Server) Definition(context.Context, *protocol.DefinitionParams) ([]protocol.Location, error) {
	panic("unimplemented")
}

// Diagnostic implements protocol.Server.
func (*Server) Diagnostic(context.Context, *string) (*string, error) {
	panic("unimplemented")
}

// DiagnosticWorkspace implements protocol.Server.
func (*Server) DiagnosticWorkspace(context.Context, *protocol.WorkspaceDiagnosticParams) (*protocol.WorkspaceDiagnosticReport, error) {
	panic("unimplemented")
}

// DidChange implements protocol.Server.
func (*Server) DidChange(context.Context, *protocol.DidChangeTextDocumentParams) error {
	panic("unimplemented")
}

// DidChangeConfiguration implements protocol.Server.
func (*Server) DidChangeConfiguration(context.Context, *protocol.DidChangeConfigurationParams) error {
	panic("unimplemented")
}

// DidChangeNotebookDocument implements protocol.Server.
func (*Server) DidChangeNotebookDocument(context.Context, *protocol.DidChangeNotebookDocumentParams) error {
	panic("unimplemented")
}

// DidChangeWatchedFiles implements protocol.Server.
func (*Server) DidChangeWatchedFiles(context.Context, *protocol.DidChangeWatchedFilesParams) error {
	panic("unimplemented")
}

// DidChangeWorkspaceFolders implements protocol.Server.
func (*Server) DidChangeWorkspaceFolders(context.Context, *protocol.DidChangeWorkspaceFoldersParams) error {
	panic("unimplemented")
}

// DidClose implements protocol.Server.
func (*Server) DidClose(context.Context, *protocol.DidCloseTextDocumentParams) error {
	panic("unimplemented")
}

// DidCloseNotebookDocument implements protocol.Server.
func (*Server) DidCloseNotebookDocument(context.Context, *protocol.DidCloseNotebookDocumentParams) error {
	panic("unimplemented")
}

// DidCreateFiles implements protocol.Server.
func (*Server) DidCreateFiles(context.Context, *protocol.CreateFilesParams) error {
	panic("unimplemented")
}

// DidDeleteFiles implements protocol.Server.
func (*Server) DidDeleteFiles(context.Context, *protocol.DeleteFilesParams) error {
	panic("unimplemented")
}

// DidOpen implements protocol.Server.
func (*Server) DidOpen(context.Context, *protocol.DidOpenTextDocumentParams) error {
	panic("unimplemented")
}

// DidOpenNotebookDocument implements protocol.Server.
func (*Server) DidOpenNotebookDocument(context.Context, *protocol.DidOpenNotebookDocumentParams) error {
	panic("unimplemented")
}

// DidRenameFiles implements protocol.Server.
func (*Server) DidRenameFiles(context.Context, *protocol.RenameFilesParams) error {
	panic("unimplemented")
}

// DidSave implements protocol.Server.
func (*Server) DidSave(context.Context, *protocol.DidSaveTextDocumentParams) error {
	panic("unimplemented")
}

// DidSaveNotebookDocument implements protocol.Server.
func (*Server) DidSaveNotebookDocument(context.Context, *protocol.DidSaveNotebookDocumentParams) error {
	panic("unimplemented")
}

// DocumentColor implements protocol.Server.
func (*Server) DocumentColor(context.Context, *protocol.DocumentColorParams) ([]protocol.ColorInformation, error) {
	panic("unimplemented")
}

// DocumentHighlight implements protocol.Server.
func (*Server) DocumentHighlight(context.Context, *protocol.DocumentHighlightParams) ([]protocol.DocumentHighlight, error) {
	panic("unimplemented")
}

// DocumentLink implements protocol.Server.
func (*Server) DocumentLink(context.Context, *protocol.DocumentLinkParams) ([]protocol.DocumentLink, error) {
	panic("unimplemented")
}

// DocumentSymbol implements protocol.Server.
func (*Server) DocumentSymbol(context.Context, *protocol.DocumentSymbolParams) ([]interface{}, error) {
	panic("unimplemented")
}

// ExecuteCommand implements protocol.Server.
func (*Server) ExecuteCommand(context.Context, *protocol.ExecuteCommandParams) (interface{}, error) {
	panic("unimplemented")
}

// Exit implements protocol.Server.
func (me *Server) Exit(ctx context.Context) error {
	return me.exit(ctx)
}

// FoldingRange implements protocol.Server.
func (*Server) FoldingRange(context.Context, *protocol.FoldingRangeParams) ([]protocol.FoldingRange, error) {
	panic("unimplemented")
}

// Formatting implements protocol.Server.
func (*Server) Formatting(context.Context, *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	panic("unimplemented")
}

// Hover implements protocol.Server.
func (*Server) Hover(context.Context, *protocol.HoverParams) (*protocol.Hover, error) {
	panic("unimplemented")
}

// Implementation implements protocol.Server.
func (*Server) Implementation(context.Context, *protocol.ImplementationParams) ([]protocol.Location, error) {
	panic("unimplemented")
}

// IncomingCalls implements protocol.Server.
func (*Server) IncomingCalls(context.Context, *protocol.CallHierarchyIncomingCallsParams) ([]protocol.CallHierarchyIncomingCall, error) {
	panic("unimplemented")
}

// Initialize implements protocol.Server.
func (me *Server) Initialize(ctx context.Context, params *protocol.ParamInitialize) (*protocol.InitializeResult, error) {
	log.Println("initializing")
	return me.initialize(ctx, params)
}

// Initialized implements protocol.Server.
func (me *Server) Initialized(ctx context.Context, params *protocol.InitializedParams) error {
	log.Println("initialized")
	return me.initialized(ctx, params)
}

// InlayHint implements protocol.Server.
func (*Server) InlayHint(context.Context, *protocol.InlayHintParams) ([]protocol.InlayHint, error) {
	panic("unimplemented")
}

// InlineCompletion implements protocol.Server.
func (*Server) InlineCompletion(context.Context, *protocol.InlineCompletionParams) (*protocol.Or_Result_textDocument_inlineCompletion, error) {
	panic("unimplemented")
}

// InlineValue implements protocol.Server.
func (*Server) InlineValue(context.Context, *protocol.InlineValueParams) ([]protocol.Or_InlineValue, error) {
	panic("unimplemented")
}

// LinkedEditingRange implements protocol.Server.
func (*Server) LinkedEditingRange(context.Context, *protocol.LinkedEditingRangeParams) (*protocol.LinkedEditingRanges, error) {
	panic("unimplemented")
}

// Moniker implements protocol.Server.
func (*Server) Moniker(context.Context, *protocol.MonikerParams) ([]protocol.Moniker, error) {
	panic("unimplemented")
}

// NonstandardRequest implements protocol.Server.
func (*Server) NonstandardRequest(ctx context.Context, method string, params interface{}) (interface{}, error) {
	panic("unimplemented")
}

// OnTypeFormatting implements protocol.Server.
func (*Server) OnTypeFormatting(context.Context, *protocol.DocumentOnTypeFormattingParams) ([]protocol.TextEdit, error) {
	panic("unimplemented")
}

// OutgoingCalls implements protocol.Server.
func (*Server) OutgoingCalls(context.Context, *protocol.CallHierarchyOutgoingCallsParams) ([]protocol.CallHierarchyOutgoingCall, error) {
	panic("unimplemented")
}

// PrepareCallHierarchy implements protocol.Server.
func (*Server) PrepareCallHierarchy(context.Context, *protocol.CallHierarchyPrepareParams) ([]protocol.CallHierarchyItem, error) {
	panic("unimplemented")
}

// PrepareRename implements protocol.Server.
func (*Server) PrepareRename(context.Context, *protocol.PrepareRenameParams) (*protocol.Msg_PrepareRename2Gn, error) {
	panic("unimplemented")
}

// PrepareTypeHierarchy implements protocol.Server.
func (*Server) PrepareTypeHierarchy(context.Context, *protocol.TypeHierarchyPrepareParams) ([]protocol.TypeHierarchyItem, error) {
	panic("unimplemented")
}

// Progress implements protocol.Server.
func (*Server) Progress(context.Context, *protocol.ProgressParams) error {
	panic("unimplemented")
}

// RangeFormatting implements protocol.Server.
func (*Server) RangeFormatting(context.Context, *protocol.DocumentRangeFormattingParams) ([]protocol.TextEdit, error) {
	panic("unimplemented")
}

// RangesFormatting implements protocol.Server.
func (*Server) RangesFormatting(context.Context, *protocol.DocumentRangesFormattingParams) ([]protocol.TextEdit, error) {
	panic("unimplemented")
}

// References implements protocol.Server.
func (*Server) References(context.Context, *protocol.ReferenceParams) ([]protocol.Location, error) {
	panic("unimplemented")
}

// Rename implements protocol.Server.
func (*Server) Rename(context.Context, *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
	panic("unimplemented")
}

// Resolve implements protocol.Server.
func (*Server) Resolve(context.Context, *protocol.InlayHint) (*protocol.InlayHint, error) {
	panic("unimplemented")
}

// ResolveCodeAction implements protocol.Server.
func (*Server) ResolveCodeAction(context.Context, *protocol.CodeAction) (*protocol.CodeAction, error) {
	panic("unimplemented")
}

// ResolveCodeLens implements protocol.Server.
func (*Server) ResolveCodeLens(context.Context, *protocol.CodeLens) (*protocol.CodeLens, error) {
	panic("unimplemented")
}

// ResolveCompletionItem implements protocol.Server.
func (*Server) ResolveCompletionItem(context.Context, *protocol.CompletionItem) (*protocol.CompletionItem, error) {
	panic("unimplemented")
}

// ResolveDocumentLink implements protocol.Server.
func (*Server) ResolveDocumentLink(context.Context, *protocol.DocumentLink) (*protocol.DocumentLink, error) {
	panic("unimplemented")
}

// ResolveWorkspaceSymbol implements protocol.Server.
func (*Server) ResolveWorkspaceSymbol(context.Context, *protocol.WorkspaceSymbol) (*protocol.WorkspaceSymbol, error) {
	panic("unimplemented")
}

// SelectionRange implements protocol.Server.
func (*Server) SelectionRange(context.Context, *protocol.SelectionRangeParams) ([]protocol.SelectionRange, error) {
	panic("unimplemented")
}

// SemanticTokensFull implements protocol.Server.
func (*Server) SemanticTokensFull(context.Context, *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	panic("unimplemented")
}

// SemanticTokensFullDelta implements protocol.Server.
func (*Server) SemanticTokensFullDelta(context.Context, *protocol.SemanticTokensDeltaParams) (interface{}, error) {
	panic("unimplemented")
}

// SemanticTokensRange implements protocol.Server.
func (*Server) SemanticTokensRange(context.Context, *protocol.SemanticTokensRangeParams) (*protocol.SemanticTokens, error) {
	panic("unimplemented")
}

// SetTrace implements protocol.Server.
func (*Server) SetTrace(context.Context, *protocol.SetTraceParams) error {
	panic("unimplemented")
}

// Shutdown implements protocol.Server.
func (me *Server) Shutdown(ctx context.Context) error {
	fmt.Println("shutting down")
	return me.shutdown(ctx)
}

// SignatureHelp implements protocol.Server.
func (*Server) SignatureHelp(context.Context, *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {
	panic("unimplemented")
}

// Subtypes implements protocol.Server.
func (*Server) Subtypes(context.Context, *protocol.TypeHierarchySubtypesParams) ([]protocol.TypeHierarchyItem, error) {
	panic("unimplemented")
}

// Supertypes implements protocol.Server.
func (*Server) Supertypes(context.Context, *protocol.TypeHierarchySupertypesParams) ([]protocol.TypeHierarchyItem, error) {
	panic("unimplemented")
}

// Symbol implements protocol.Server.
func (*Server) Symbol(context.Context, *protocol.WorkspaceSymbolParams) ([]protocol.SymbolInformation, error) {
	panic("unimplemented")
}

// TypeDefinition implements protocol.Server.
func (*Server) TypeDefinition(context.Context, *protocol.TypeDefinitionParams) ([]protocol.Location, error) {
	panic("unimplemented")
}

// WillCreateFiles implements protocol.Server.
func (*Server) WillCreateFiles(context.Context, *protocol.CreateFilesParams) (*protocol.WorkspaceEdit, error) {
	panic("unimplemented")
}

// WillDeleteFiles implements protocol.Server.
func (*Server) WillDeleteFiles(context.Context, *protocol.DeleteFilesParams) (*protocol.WorkspaceEdit, error) {
	panic("unimplemented")
}

// WillRenameFiles implements protocol.Server.
func (*Server) WillRenameFiles(context.Context, *protocol.RenameFilesParams) (*protocol.WorkspaceEdit, error) {
	panic("unimplemented")
}

// WillSave implements protocol.Server.
func (*Server) WillSave(context.Context, *protocol.WillSaveTextDocumentParams) error {
	panic("unimplemented")
}

// WillSaveWaitUntil implements protocol.Server.
func (*Server) WillSaveWaitUntil(context.Context, *protocol.WillSaveTextDocumentParams) ([]protocol.TextEdit, error) {
	panic("unimplemented")
}

// WorkDoneProgressCancel implements protocol.Server.
func (*Server) WorkDoneProgressCancel(context.Context, *protocol.WorkDoneProgressCancelParams) error {
	panic("unimplemented")
}
