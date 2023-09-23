package server

import (
	"context"

	"github.com/walteh/retab/gen/gopls"
)

type Server struct {
}

var _ gopls.Server = (*Server)(nil)

// CodeAction implements gopls.Server.
func (*Server) CodeAction(context.Context, *gopls.CodeActionParams) ([]gopls.CodeAction, error) {
	panic("unimplemented")
}

// CodeLens implements gopls.Server.
func (*Server) CodeLens(context.Context, *gopls.CodeLensParams) ([]gopls.CodeLens, error) {
	panic("unimplemented")
}

// ColorPresentation implements gopls.Server.
func (*Server) ColorPresentation(context.Context, *gopls.ColorPresentationParams) ([]gopls.ColorPresentation, error) {
	panic("unimplemented")
}

// Completion implements gopls.Server.
func (*Server) Completion(context.Context, *gopls.CompletionParams) (*gopls.CompletionList, error) {
	panic("unimplemented")
}

// Declaration implements gopls.Server.
func (*Server) Declaration(context.Context, *gopls.DeclarationParams) (*gopls.Or_textDocument_declaration, error) {
	panic("unimplemented")
}

// Definition implements gopls.Server.
func (*Server) Definition(context.Context, *gopls.DefinitionParams) ([]gopls.Location, error) {
	panic("unimplemented")
}

// Diagnostic implements gopls.Server.
func (*Server) Diagnostic(context.Context, *string) (*string, error) {
	panic("unimplemented")
}

// DiagnosticWorkspace implements gopls.Server.
func (*Server) DiagnosticWorkspace(context.Context, *gopls.WorkspaceDiagnosticParams) (*gopls.WorkspaceDiagnosticReport, error) {
	panic("unimplemented")
}

// DidChange implements gopls.Server.
func (*Server) DidChange(context.Context, *gopls.DidChangeTextDocumentParams) error {
	panic("unimplemented")
}

// DidChangeConfiguration implements gopls.Server.
func (*Server) DidChangeConfiguration(context.Context, *gopls.DidChangeConfigurationParams) error {
	panic("unimplemented")
}

// DidChangeNotebookDocument implements gopls.Server.
func (*Server) DidChangeNotebookDocument(context.Context, *gopls.DidChangeNotebookDocumentParams) error {
	panic("unimplemented")
}

// DidChangeWatchedFiles implements gopls.Server.
func (*Server) DidChangeWatchedFiles(context.Context, *gopls.DidChangeWatchedFilesParams) error {
	panic("unimplemented")
}

// DidChangeWorkspaceFolders implements gopls.Server.
func (*Server) DidChangeWorkspaceFolders(context.Context, *gopls.DidChangeWorkspaceFoldersParams) error {
	panic("unimplemented")
}

// DidClose implements gopls.Server.
func (*Server) DidClose(context.Context, *gopls.DidCloseTextDocumentParams) error {
	panic("unimplemented")
}

// DidCloseNotebookDocument implements gopls.Server.
func (*Server) DidCloseNotebookDocument(context.Context, *gopls.DidCloseNotebookDocumentParams) error {
	panic("unimplemented")
}

// DidCreateFiles implements gopls.Server.
func (*Server) DidCreateFiles(context.Context, *gopls.CreateFilesParams) error {
	panic("unimplemented")
}

// DidDeleteFiles implements gopls.Server.
func (*Server) DidDeleteFiles(context.Context, *gopls.DeleteFilesParams) error {
	panic("unimplemented")
}

// DidOpen implements gopls.Server.
func (*Server) DidOpen(context.Context, *gopls.DidOpenTextDocumentParams) error {
	panic("unimplemented")
}

// DidOpenNotebookDocument implements gopls.Server.
func (*Server) DidOpenNotebookDocument(context.Context, *gopls.DidOpenNotebookDocumentParams) error {
	panic("unimplemented")
}

// DidRenameFiles implements gopls.Server.
func (*Server) DidRenameFiles(context.Context, *gopls.RenameFilesParams) error {
	panic("unimplemented")
}

// DidSave implements gopls.Server.
func (*Server) DidSave(context.Context, *gopls.DidSaveTextDocumentParams) error {
	panic("unimplemented")
}

// DidSaveNotebookDocument implements gopls.Server.
func (*Server) DidSaveNotebookDocument(context.Context, *gopls.DidSaveNotebookDocumentParams) error {
	panic("unimplemented")
}

// DocumentColor implements gopls.Server.
func (*Server) DocumentColor(context.Context, *gopls.DocumentColorParams) ([]gopls.ColorInformation, error) {
	panic("unimplemented")
}

// DocumentHighlight implements gopls.Server.
func (*Server) DocumentHighlight(context.Context, *gopls.DocumentHighlightParams) ([]gopls.DocumentHighlight, error) {
	panic("unimplemented")
}

// DocumentLink implements gopls.Server.
func (*Server) DocumentLink(context.Context, *gopls.DocumentLinkParams) ([]gopls.DocumentLink, error) {
	panic("unimplemented")
}

// DocumentSymbol implements gopls.Server.
func (*Server) DocumentSymbol(context.Context, *gopls.DocumentSymbolParams) ([]interface{}, error) {
	panic("unimplemented")
}

// ExecuteCommand implements gopls.Server.
func (*Server) ExecuteCommand(context.Context, *gopls.ExecuteCommandParams) (interface{}, error) {
	panic("unimplemented")
}

// Exit implements gopls.Server.
func (*Server) Exit(context.Context) error {
	panic("unimplemented")
}

// FoldingRange implements gopls.Server.
func (*Server) FoldingRange(context.Context, *gopls.FoldingRangeParams) ([]gopls.FoldingRange, error) {
	panic("unimplemented")
}

// Formatting implements gopls.Server.
func (*Server) Formatting(context.Context, *gopls.DocumentFormattingParams) ([]gopls.TextEdit, error) {
	panic("unimplemented")
}

// Hover implements gopls.Server.
func (*Server) Hover(context.Context, *gopls.HoverParams) (*gopls.Hover, error) {
	panic("unimplemented")
}

// Implementation implements gopls.Server.
func (*Server) Implementation(context.Context, *gopls.ImplementationParams) ([]gopls.Location, error) {
	panic("unimplemented")
}

// IncomingCalls implements gopls.Server.
func (*Server) IncomingCalls(context.Context, *gopls.CallHierarchyIncomingCallsParams) ([]gopls.CallHierarchyIncomingCall, error) {
	panic("unimplemented")
}

// Initialize implements gopls.Server.
func (*Server) Initialize(context.Context, *gopls.ParamInitialize) (*gopls.InitializeResult, error) {
	panic("unimplemented")
}

// Initialized implements gopls.Server.
func (*Server) Initialized(context.Context, *gopls.InitializedParams) error {
	panic("unimplemented")
}

// InlayHint implements gopls.Server.
func (*Server) InlayHint(context.Context, *gopls.InlayHintParams) ([]gopls.InlayHint, error) {
	panic("unimplemented")
}

// InlineCompletion implements gopls.Server.
func (*Server) InlineCompletion(context.Context, *gopls.InlineCompletionParams) (*gopls.Or_Result_textDocument_inlineCompletion, error) {
	panic("unimplemented")
}

// InlineValue implements gopls.Server.
func (*Server) InlineValue(context.Context, *gopls.InlineValueParams) ([]gopls.Or_InlineValue, error) {
	panic("unimplemented")
}

// LinkedEditingRange implements gopls.Server.
func (*Server) LinkedEditingRange(context.Context, *gopls.LinkedEditingRangeParams) (*gopls.LinkedEditingRanges, error) {
	panic("unimplemented")
}

// Moniker implements gopls.Server.
func (*Server) Moniker(context.Context, *gopls.MonikerParams) ([]gopls.Moniker, error) {
	panic("unimplemented")
}

// NonstandardRequest implements gopls.Server.
func (*Server) NonstandardRequest(ctx context.Context, method string, params interface{}) (interface{}, error) {
	panic("unimplemented")
}

// OnTypeFormatting implements gopls.Server.
func (*Server) OnTypeFormatting(context.Context, *gopls.DocumentOnTypeFormattingParams) ([]gopls.TextEdit, error) {
	panic("unimplemented")
}

// OutgoingCalls implements gopls.Server.
func (*Server) OutgoingCalls(context.Context, *gopls.CallHierarchyOutgoingCallsParams) ([]gopls.CallHierarchyOutgoingCall, error) {
	panic("unimplemented")
}

// PrepareCallHierarchy implements gopls.Server.
func (*Server) PrepareCallHierarchy(context.Context, *gopls.CallHierarchyPrepareParams) ([]gopls.CallHierarchyItem, error) {
	panic("unimplemented")
}

// PrepareRename implements gopls.Server.
func (*Server) PrepareRename(context.Context, *gopls.PrepareRenameParams) (*gopls.Msg_PrepareRename2Gn, error) {
	panic("unimplemented")
}

// PrepareTypeHierarchy implements gopls.Server.
func (*Server) PrepareTypeHierarchy(context.Context, *gopls.TypeHierarchyPrepareParams) ([]gopls.TypeHierarchyItem, error) {
	panic("unimplemented")
}

// Progress implements gopls.Server.
func (*Server) Progress(context.Context, *gopls.ProgressParams) error {
	panic("unimplemented")
}

// RangeFormatting implements gopls.Server.
func (*Server) RangeFormatting(context.Context, *gopls.DocumentRangeFormattingParams) ([]gopls.TextEdit, error) {
	panic("unimplemented")
}

// RangesFormatting implements gopls.Server.
func (*Server) RangesFormatting(context.Context, *gopls.DocumentRangesFormattingParams) ([]gopls.TextEdit, error) {
	panic("unimplemented")
}

// References implements gopls.Server.
func (*Server) References(context.Context, *gopls.ReferenceParams) ([]gopls.Location, error) {
	panic("unimplemented")
}

// Rename implements gopls.Server.
func (*Server) Rename(context.Context, *gopls.RenameParams) (*gopls.WorkspaceEdit, error) {
	panic("unimplemented")
}

// Resolve implements gopls.Server.
func (*Server) Resolve(context.Context, *gopls.InlayHint) (*gopls.InlayHint, error) {
	panic("unimplemented")
}

// ResolveCodeAction implements gopls.Server.
func (*Server) ResolveCodeAction(context.Context, *gopls.CodeAction) (*gopls.CodeAction, error) {
	panic("unimplemented")
}

// ResolveCodeLens implements gopls.Server.
func (*Server) ResolveCodeLens(context.Context, *gopls.CodeLens) (*gopls.CodeLens, error) {
	panic("unimplemented")
}

// ResolveCompletionItem implements gopls.Server.
func (*Server) ResolveCompletionItem(context.Context, *gopls.CompletionItem) (*gopls.CompletionItem, error) {
	panic("unimplemented")
}

// ResolveDocumentLink implements gopls.Server.
func (*Server) ResolveDocumentLink(context.Context, *gopls.DocumentLink) (*gopls.DocumentLink, error) {
	panic("unimplemented")
}

// ResolveWorkspaceSymbol implements gopls.Server.
func (*Server) ResolveWorkspaceSymbol(context.Context, *gopls.WorkspaceSymbol) (*gopls.WorkspaceSymbol, error) {
	panic("unimplemented")
}

// SelectionRange implements gopls.Server.
func (*Server) SelectionRange(context.Context, *gopls.SelectionRangeParams) ([]gopls.SelectionRange, error) {
	panic("unimplemented")
}

// SemanticTokensFull implements gopls.Server.
func (*Server) SemanticTokensFull(context.Context, *gopls.SemanticTokensParams) (*gopls.SemanticTokens, error) {
	panic("unimplemented")
}

// SemanticTokensFullDelta implements gopls.Server.
func (*Server) SemanticTokensFullDelta(context.Context, *gopls.SemanticTokensDeltaParams) (interface{}, error) {
	panic("unimplemented")
}

// SemanticTokensRange implements gopls.Server.
func (*Server) SemanticTokensRange(context.Context, *gopls.SemanticTokensRangeParams) (*gopls.SemanticTokens, error) {
	panic("unimplemented")
}

// SetTrace implements gopls.Server.
func (*Server) SetTrace(context.Context, *gopls.SetTraceParams) error {
	panic("unimplemented")
}

// Shutdown implements gopls.Server.
func (*Server) Shutdown(context.Context) error {
	panic("unimplemented")
}

// SignatureHelp implements gopls.Server.
func (*Server) SignatureHelp(context.Context, *gopls.SignatureHelpParams) (*gopls.SignatureHelp, error) {
	panic("unimplemented")
}

// Subtypes implements gopls.Server.
func (*Server) Subtypes(context.Context, *gopls.TypeHierarchySubtypesParams) ([]gopls.TypeHierarchyItem, error) {
	panic("unimplemented")
}

// Supertypes implements gopls.Server.
func (*Server) Supertypes(context.Context, *gopls.TypeHierarchySupertypesParams) ([]gopls.TypeHierarchyItem, error) {
	panic("unimplemented")
}

// Symbol implements gopls.Server.
func (*Server) Symbol(context.Context, *gopls.WorkspaceSymbolParams) ([]gopls.SymbolInformation, error) {
	panic("unimplemented")
}

// TypeDefinition implements gopls.Server.
func (*Server) TypeDefinition(context.Context, *gopls.TypeDefinitionParams) ([]gopls.Location, error) {
	panic("unimplemented")
}

// WillCreateFiles implements gopls.Server.
func (*Server) WillCreateFiles(context.Context, *gopls.CreateFilesParams) (*gopls.WorkspaceEdit, error) {
	panic("unimplemented")
}

// WillDeleteFiles implements gopls.Server.
func (*Server) WillDeleteFiles(context.Context, *gopls.DeleteFilesParams) (*gopls.WorkspaceEdit, error) {
	panic("unimplemented")
}

// WillRenameFiles implements gopls.Server.
func (*Server) WillRenameFiles(context.Context, *gopls.RenameFilesParams) (*gopls.WorkspaceEdit, error) {
	panic("unimplemented")
}

// WillSave implements gopls.Server.
func (*Server) WillSave(context.Context, *gopls.WillSaveTextDocumentParams) error {
	panic("unimplemented")
}

// WillSaveWaitUntil implements gopls.Server.
func (*Server) WillSaveWaitUntil(context.Context, *gopls.WillSaveTextDocumentParams) ([]gopls.TextEdit, error) {
	panic("unimplemented")
}

// WorkDoneProgressCancel implements gopls.Server.
func (*Server) WorkDoneProgressCancel(context.Context, *gopls.WorkDoneProgressCancelParams) error {
	panic("unimplemented")
}
