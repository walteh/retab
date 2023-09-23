package server

import (
	"context"

	"github.com/walteh/retab/gen/gopls/protocol"
)

var _ protocol.Server = (*Server)(nil)

// CodeAction implements gopls.Server.
func (*Server) CodeAction(context.Context, *protocol.CodeActionParams) ([]protocol.CodeAction, error) {
	panic("unimplemented")
}

// CodeLens implements gopls.Server.
func (*Server) CodeLens(context.Context, *protocol.CodeLensParams) ([]protocol.CodeLens, error) {
	panic("unimplemented")
}

// ColorPresentation implements gopls.Server.
func (*Server) ColorPresentation(context.Context, *protocol.ColorPresentationParams) ([]protocol.ColorPresentation, error) {
	panic("unimplemented")
}

// Completion implements gopls.Server.
func (*Server) Completion(context.Context, *protocol.CompletionParams) (*protocol.CompletionList, error) {
	panic("unimplemented")
}

// Declaration implements gopls.Server.
func (*Server) Declaration(context.Context, *protocol.DeclarationParams) (*protocol.Or_textDocument_declaration, error) {
	panic("unimplemented")
}

// Definition implements gopls.Server.
func (*Server) Definition(context.Context, *protocol.DefinitionParams) ([]protocol.Location, error) {
	panic("unimplemented")
}

// Diagnostic implements gopls.Server.
func (*Server) Diagnostic(context.Context, *string) (*string, error) {
	panic("unimplemented")
}

// DiagnosticWorkspace implements gopls.Server.
func (*Server) DiagnosticWorkspace(context.Context, *protocol.WorkspaceDiagnosticParams) (*protocol.WorkspaceDiagnosticReport, error) {
	panic("unimplemented")
}

// DidChange implements gopls.Server.
func (*Server) DidChange(context.Context, *protocol.DidChangeTextDocumentParams) error {
	panic("unimplemented")
}

// DidChangeConfiguration implements gopls.Server.
func (*Server) DidChangeConfiguration(context.Context, *protocol.DidChangeConfigurationParams) error {
	panic("unimplemented")
}

// DidChangeNotebookDocument implements gopls.Server.
func (*Server) DidChangeNotebookDocument(context.Context, *protocol.DidChangeNotebookDocumentParams) error {
	panic("unimplemented")
}

// DidChangeWatchedFiles implements gopls.Server.
func (*Server) DidChangeWatchedFiles(context.Context, *protocol.DidChangeWatchedFilesParams) error {
	panic("unimplemented")
}

// DidChangeWorkspaceFolders implements gopls.Server.
func (*Server) DidChangeWorkspaceFolders(context.Context, *protocol.DidChangeWorkspaceFoldersParams) error {
	panic("unimplemented")
}

// DidClose implements gopls.Server.
func (*Server) DidClose(context.Context, *protocol.DidCloseTextDocumentParams) error {
	panic("unimplemented")
}

// DidCloseNotebookDocument implements gopls.Server.
func (*Server) DidCloseNotebookDocument(context.Context, *protocol.DidCloseNotebookDocumentParams) error {
	panic("unimplemented")
}

// DidCreateFiles implements gopls.Server.
func (*Server) DidCreateFiles(context.Context, *protocol.CreateFilesParams) error {
	panic("unimplemented")
}

// DidDeleteFiles implements gopls.Server.
func (*Server) DidDeleteFiles(context.Context, *protocol.DeleteFilesParams) error {
	panic("unimplemented")
}

// DidOpen implements gopls.Server.
func (*Server) DidOpen(context.Context, *protocol.DidOpenTextDocumentParams) error {
	panic("unimplemented")
}

// DidOpenNotebookDocument implements gopls.Server.
func (*Server) DidOpenNotebookDocument(context.Context, *protocol.DidOpenNotebookDocumentParams) error {
	panic("unimplemented")
}

// DidRenameFiles implements gopls.Server.
func (*Server) DidRenameFiles(context.Context, *protocol.RenameFilesParams) error {
	panic("unimplemented")
}

// DidSave implements gopls.Server.
func (*Server) DidSave(context.Context, *protocol.DidSaveTextDocumentParams) error {
	panic("unimplemented")
}

// DidSaveNotebookDocument implements gopls.Server.
func (*Server) DidSaveNotebookDocument(context.Context, *protocol.DidSaveNotebookDocumentParams) error {
	panic("unimplemented")
}

// DocumentColor implements gopls.Server.
func (*Server) DocumentColor(context.Context, *protocol.DocumentColorParams) ([]protocol.ColorInformation, error) {
	panic("unimplemented")
}

// DocumentHighlight implements gopls.Server.
func (*Server) DocumentHighlight(context.Context, *protocol.DocumentHighlightParams) ([]protocol.DocumentHighlight, error) {
	panic("unimplemented")
}

// DocumentLink implements gopls.Server.
func (*Server) DocumentLink(context.Context, *protocol.DocumentLinkParams) ([]protocol.DocumentLink, error) {
	panic("unimplemented")
}

// DocumentSymbol implements gopls.Server.
func (*Server) DocumentSymbol(context.Context, *protocol.DocumentSymbolParams) ([]interface{}, error) {
	panic("unimplemented")
}

// ExecuteCommand implements gopls.Server.
func (*Server) ExecuteCommand(context.Context, *protocol.ExecuteCommandParams) (interface{}, error) {
	panic("unimplemented")
}

// Exit implements gopls.Server.
func (*Server) Exit(context.Context) error {
	panic("unimplemented")
}

// FoldingRange implements gopls.Server.
func (*Server) FoldingRange(context.Context, *protocol.FoldingRangeParams) ([]protocol.FoldingRange, error) {
	panic("unimplemented")
}

// Formatting implements gopls.Server.
func (*Server) Formatting(context.Context, *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	panic("unimplemented")
}

// Hover implements gopls.Server.
func (*Server) Hover(context.Context, *protocol.HoverParams) (*protocol.Hover, error) {
	panic("unimplemented")
}

// Implementation implements gopls.Server.
func (*Server) Implementation(context.Context, *protocol.ImplementationParams) ([]protocol.Location, error) {
	panic("unimplemented")
}

// IncomingCalls implements gopls.Server.
func (*Server) IncomingCalls(context.Context, *protocol.CallHierarchyIncomingCallsParams) ([]protocol.CallHierarchyIncomingCall, error) {
	panic("unimplemented")
}

// Initialize implements gopls.Server.
func (*Server) Initialize(context.Context, *protocol.ParamInitialize) (*protocol.InitializeResult, error) {
	panic("unimplemented")
}

// Initialized implements gopls.Server.
func (*Server) Initialized(context.Context, *protocol.InitializedParams) error {
	panic("unimplemented")
}

// InlayHint implements gopls.Server.
func (*Server) InlayHint(context.Context, *protocol.InlayHintParams) ([]protocol.InlayHint, error) {
	panic("unimplemented")
}

// InlineCompletion implements gopls.Server.
func (*Server) InlineCompletion(context.Context, *protocol.InlineCompletionParams) (*protocol.Or_Result_textDocument_inlineCompletion, error) {
	panic("unimplemented")
}

// InlineValue implements gopls.Server.
func (*Server) InlineValue(context.Context, *protocol.InlineValueParams) ([]protocol.Or_InlineValue, error) {
	panic("unimplemented")
}

// LinkedEditingRange implements gopls.Server.
func (*Server) LinkedEditingRange(context.Context, *protocol.LinkedEditingRangeParams) (*protocol.LinkedEditingRanges, error) {
	panic("unimplemented")
}

// Moniker implements gopls.Server.
func (*Server) Moniker(context.Context, *protocol.MonikerParams) ([]protocol.Moniker, error) {
	panic("unimplemented")
}

// NonstandardRequest implements gopls.Server.
func (*Server) NonstandardRequest(ctx context.Context, method string, params interface{}) (interface{}, error) {
	panic("unimplemented")
}

// OnTypeFormatting implements gopls.Server.
func (*Server) OnTypeFormatting(context.Context, *protocol.DocumentOnTypeFormattingParams) ([]protocol.TextEdit, error) {
	panic("unimplemented")
}

// OutgoingCalls implements gopls.Server.
func (*Server) OutgoingCalls(context.Context, *protocol.CallHierarchyOutgoingCallsParams) ([]protocol.CallHierarchyOutgoingCall, error) {
	panic("unimplemented")
}

// PrepareCallHierarchy implements gopls.Server.
func (*Server) PrepareCallHierarchy(context.Context, *protocol.CallHierarchyPrepareParams) ([]protocol.CallHierarchyItem, error) {
	panic("unimplemented")
}

// PrepareRename implements gopls.Server.
func (*Server) PrepareRename(context.Context, *protocol.PrepareRenameParams) (*protocol.Msg_PrepareRename2Gn, error) {
	panic("unimplemented")
}

// PrepareTypeHierarchy implements gopls.Server.
func (*Server) PrepareTypeHierarchy(context.Context, *protocol.TypeHierarchyPrepareParams) ([]protocol.TypeHierarchyItem, error) {
	panic("unimplemented")
}

// Progress implements gopls.Server.
func (*Server) Progress(context.Context, *protocol.ProgressParams) error {
	panic("unimplemented")
}

// RangeFormatting implements gopls.Server.
func (*Server) RangeFormatting(context.Context, *protocol.DocumentRangeFormattingParams) ([]protocol.TextEdit, error) {
	panic("unimplemented")
}

// RangesFormatting implements gopls.Server.
func (*Server) RangesFormatting(context.Context, *protocol.DocumentRangesFormattingParams) ([]protocol.TextEdit, error) {
	panic("unimplemented")
}

// References implements gopls.Server.
func (*Server) References(context.Context, *protocol.ReferenceParams) ([]protocol.Location, error) {
	panic("unimplemented")
}

// Rename implements gopls.Server.
func (*Server) Rename(context.Context, *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
	panic("unimplemented")
}

// Resolve implements gopls.Server.
func (*Server) Resolve(context.Context, *protocol.InlayHint) (*protocol.InlayHint, error) {
	panic("unimplemented")
}

// ResolveCodeAction implements gopls.Server.
func (*Server) ResolveCodeAction(context.Context, *protocol.CodeAction) (*protocol.CodeAction, error) {
	panic("unimplemented")
}

// ResolveCodeLens implements gopls.Server.
func (*Server) ResolveCodeLens(context.Context, *protocol.CodeLens) (*protocol.CodeLens, error) {
	panic("unimplemented")
}

// ResolveCompletionItem implements gopls.Server.
func (*Server) ResolveCompletionItem(context.Context, *protocol.CompletionItem) (*protocol.CompletionItem, error) {
	panic("unimplemented")
}

// ResolveDocumentLink implements gopls.Server.
func (*Server) ResolveDocumentLink(context.Context, *protocol.DocumentLink) (*protocol.DocumentLink, error) {
	panic("unimplemented")
}

// ResolveWorkspaceSymbol implements gopls.Server.
func (*Server) ResolveWorkspaceSymbol(context.Context, *protocol.WorkspaceSymbol) (*protocol.WorkspaceSymbol, error) {
	panic("unimplemented")
}

// SelectionRange implements gopls.Server.
func (*Server) SelectionRange(context.Context, *protocol.SelectionRangeParams) ([]protocol.SelectionRange, error) {
	panic("unimplemented")
}

// SemanticTokensFull implements gopls.Server.
func (*Server) SemanticTokensFull(context.Context, *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	panic("unimplemented")
}

// SemanticTokensFullDelta implements gopls.Server.
func (*Server) SemanticTokensFullDelta(context.Context, *protocol.SemanticTokensDeltaParams) (interface{}, error) {
	panic("unimplemented")
}

// SemanticTokensRange implements gopls.Server.
func (*Server) SemanticTokensRange(context.Context, *protocol.SemanticTokensRangeParams) (*protocol.SemanticTokens, error) {
	panic("unimplemented")
}

// SetTrace implements gopls.Server.
func (*Server) SetTrace(context.Context, *protocol.SetTraceParams) error {
	panic("unimplemented")
}

// Shutdown implements gopls.Server.
func (*Server) Shutdown(context.Context) error {
	panic("unimplemented")
}

// SignatureHelp implements gopls.Server.
func (*Server) SignatureHelp(context.Context, *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {
	panic("unimplemented")
}

// Subtypes implements gopls.Server.
func (*Server) Subtypes(context.Context, *protocol.TypeHierarchySubtypesParams) ([]protocol.TypeHierarchyItem, error) {
	panic("unimplemented")
}

// Supertypes implements gopls.Server.
func (*Server) Supertypes(context.Context, *protocol.TypeHierarchySupertypesParams) ([]protocol.TypeHierarchyItem, error) {
	panic("unimplemented")
}

// Symbol implements gopls.Server.
func (*Server) Symbol(context.Context, *protocol.WorkspaceSymbolParams) ([]protocol.SymbolInformation, error) {
	panic("unimplemented")
}

// TypeDefinition implements gopls.Server.
func (*Server) TypeDefinition(context.Context, *protocol.TypeDefinitionParams) ([]protocol.Location, error) {
	panic("unimplemented")
}

// WillCreateFiles implements gopls.Server.
func (*Server) WillCreateFiles(context.Context, *protocol.CreateFilesParams) (*protocol.WorkspaceEdit, error) {
	panic("unimplemented")
}

// WillDeleteFiles implements gopls.Server.
func (*Server) WillDeleteFiles(context.Context, *protocol.DeleteFilesParams) (*protocol.WorkspaceEdit, error) {
	panic("unimplemented")
}

// WillRenameFiles implements gopls.Server.
func (*Server) WillRenameFiles(context.Context, *protocol.RenameFilesParams) (*protocol.WorkspaceEdit, error) {
	panic("unimplemented")
}

// WillSave implements gopls.Server.
func (*Server) WillSave(context.Context, *protocol.WillSaveTextDocumentParams) error {
	panic("unimplemented")
}

// WillSaveWaitUntil implements gopls.Server.
func (*Server) WillSaveWaitUntil(context.Context, *protocol.WillSaveTextDocumentParams) ([]protocol.TextEdit, error) {
	panic("unimplemented")
}

// WorkDoneProgressCancel implements gopls.Server.
func (*Server) WorkDoneProgressCancel(context.Context, *protocol.WorkDoneProgressCancelParams) error {
	panic("unimplemented")
}
