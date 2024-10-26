package main

import (
	"context"

	"go.lsp.dev/protocol"
)

// BaseServer provides no-op implementations for all protocol.Server methods
type BaseServer struct{}

func (s *BaseServer) CodeLens(ctx context.Context, params *protocol.CodeLensParams) ([]protocol.CodeLens, error) {
	return nil, nil
}

func (s *BaseServer) CodeLensRefresh(ctx context.Context) error {
	return nil
}

func (s *BaseServer) CodeLensResolve(ctx context.Context, params *protocol.CodeLens) (*protocol.CodeLens, error) {
	return nil, nil
}

func (s *BaseServer) Declaration(ctx context.Context, params *protocol.DeclarationParams) ([]protocol.Location, error) {
	return nil, nil
}

func (s *BaseServer) Definition(ctx context.Context, params *protocol.DefinitionParams) ([]protocol.Location, error) {
	return nil, nil
}

func (s *BaseServer) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	return nil
}

func (s *BaseServer) DidSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) error {
	return nil
}

func (s *BaseServer) DocumentColor(ctx context.Context, params *protocol.DocumentColorParams) ([]protocol.ColorInformation, error) {
	return nil, nil
}

func (s *BaseServer) DocumentHighlight(ctx context.Context, params *protocol.DocumentHighlightParams) ([]protocol.DocumentHighlight, error) {
	return nil, nil
}

func (s *BaseServer) DocumentLink(ctx context.Context, params *protocol.DocumentLinkParams) ([]protocol.DocumentLink, error) {
	return nil, nil
}

func (s *BaseServer) DocumentSymbol(ctx context.Context, params *protocol.DocumentSymbolParams) ([]interface{}, error) {
	return nil, nil
}

func (s *BaseServer) FoldingRange(ctx context.Context, params *protocol.FoldingRangeParams) ([]protocol.FoldingRange, error) {
	return []protocol.FoldingRange{
		{
			StartLine:      0,
			StartCharacter: 0,
			EndLine:        10,
			EndCharacter:   0,
			Kind:           "region",
		},
	}, nil
}
func (s *BaseServer) FoldingRanges(ctx context.Context, params *protocol.FoldingRangeParams) (result []protocol.FoldingRange, err error) {
	return nil, nil
}

func (s *BaseServer) Formatting(ctx context.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	return nil, nil
}

func (s *BaseServer) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	return nil, nil
}

func (s *BaseServer) Implementation(ctx context.Context, params *protocol.ImplementationParams) ([]protocol.Location, error) {
	return nil, nil
}

func (s *BaseServer) OnTypeFormatting(ctx context.Context, params *protocol.DocumentOnTypeFormattingParams) ([]protocol.TextEdit, error) {
	return []protocol.TextEdit{
		{
			Range: protocol.Range{
				Start: protocol.Position{Line: params.Position.Line, Character: 0},
				End:   protocol.Position{Line: params.Position.Line, Character: params.Position.Character},
			},
			NewText: "// Formatted on type: " + params.Ch,
		},
	}, nil
}

func (s *BaseServer) PrepareRename(ctx context.Context, params *protocol.PrepareRenameParams) (*protocol.Range, error) {
	return nil, nil
}

func (s *BaseServer) RangeFormatting(ctx context.Context, params *protocol.DocumentRangeFormattingParams) ([]protocol.TextEdit, error) {
	return nil, nil
}

func (s *BaseServer) References(ctx context.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
	return nil, nil
}

func (s *BaseServer) Rename(ctx context.Context, params *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
	return nil, nil
}

func (s *BaseServer) SelectionRange(ctx context.Context, params *protocol.SelectionRangeParams) ([]protocol.SelectionRange, error) {
	return nil, nil
}

func (s *BaseServer) SignatureHelp(ctx context.Context, params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {
	return nil, nil
}

func (s *BaseServer) TypeDefinition(ctx context.Context, params *protocol.TypeDefinitionParams) ([]protocol.Location, error) {
	return nil, nil
}

func (s *BaseServer) WillSave(ctx context.Context, params *protocol.WillSaveTextDocumentParams) error {
	return nil
}

func (s *BaseServer) WillSaveWaitUntil(ctx context.Context, params *protocol.WillSaveTextDocumentParams) ([]protocol.TextEdit, error) {
	return nil, nil
}

func (s *BaseServer) WorkDoneProgressCancel(ctx context.Context, params *protocol.WorkDoneProgressCancelParams) error {
	return nil
}

func (s *BaseServer) Progress(ctx context.Context, params *protocol.ProgressParams) error {
	return nil
}

func (s *BaseServer) LogTrace(ctx context.Context, params *protocol.LogTraceParams) error {
	return nil
}

func (s *BaseServer) SetTrace(ctx context.Context, params *protocol.SetTraceParams) error {
	return nil
}

func (s *BaseServer) LinkedEditingRange(ctx context.Context, params *protocol.LinkedEditingRangeParams) (*protocol.LinkedEditingRanges, error) {
	return &protocol.LinkedEditingRanges{}, nil
}

func (s *BaseServer) CompletionResolve(ctx context.Context, params *protocol.CompletionItem) (*protocol.CompletionItem, error) {
	return params, nil
}

func (s *BaseServer) DidChangeConfiguration(ctx context.Context, params *protocol.DidChangeConfigurationParams) error {
	return nil
}

func (s *BaseServer) DidChangeWatchedFiles(ctx context.Context, params *protocol.DidChangeWatchedFilesParams) error {
	return nil
}

func (s *BaseServer) DidChangeWorkspaceFolders(ctx context.Context, params *protocol.DidChangeWorkspaceFoldersParams) error {
	return nil
}

func (s *BaseServer) DocumentLinkResolve(ctx context.Context, params *protocol.DocumentLink) (*protocol.DocumentLink, error) {
	return params, nil
}

func (s *BaseServer) ExecuteCommand(ctx context.Context, params *protocol.ExecuteCommandParams) (interface{}, error) {
	return nil, nil
}

func (s *BaseServer) ShowDocument(ctx context.Context, params *protocol.ShowDocumentParams) (*protocol.ShowDocumentResult, error) {
	return &protocol.ShowDocumentResult{Success: true}, nil
}

func (s *BaseServer) WillCreateFiles(ctx context.Context, params *protocol.CreateFilesParams) (*protocol.WorkspaceEdit, error) {
	return &protocol.WorkspaceEdit{}, nil
}

func (s *BaseServer) DidCreateFiles(ctx context.Context, params *protocol.CreateFilesParams) error {
	return nil
}

func (s *BaseServer) WillRenameFiles(ctx context.Context, params *protocol.RenameFilesParams) (*protocol.WorkspaceEdit, error) {
	return &protocol.WorkspaceEdit{}, nil
}

func (s *BaseServer) DidRenameFiles(ctx context.Context, params *protocol.RenameFilesParams) error {
	return nil
}

func (s *BaseServer) WillDeleteFiles(ctx context.Context, params *protocol.DeleteFilesParams) (*protocol.WorkspaceEdit, error) {
	return &protocol.WorkspaceEdit{}, nil
}

func (s *BaseServer) DidDeleteFiles(ctx context.Context, params *protocol.DeleteFilesParams) error {
	return nil
}

func (s *BaseServer) Request(ctx context.Context, method string, params interface{}) (interface{}, error) {
	return nil, nil
}

func (s *BaseServer) Symbols(ctx context.Context, params *protocol.WorkspaceSymbolParams) ([]protocol.SymbolInformation, error) {
	return []protocol.SymbolInformation{}, nil
}

func (s *BaseServer) PrepareCallHierarchy(ctx context.Context, params *protocol.CallHierarchyPrepareParams) ([]protocol.CallHierarchyItem, error) {
	return []protocol.CallHierarchyItem{}, nil
}

func (s *BaseServer) IncomingCalls(ctx context.Context, params *protocol.CallHierarchyIncomingCallsParams) ([]protocol.CallHierarchyIncomingCall, error) {
	return []protocol.CallHierarchyIncomingCall{}, nil
}

func (s *BaseServer) OutgoingCalls(ctx context.Context, params *protocol.CallHierarchyOutgoingCallsParams) ([]protocol.CallHierarchyOutgoingCall, error) {
	return []protocol.CallHierarchyOutgoingCall{}, nil
}

func (s *BaseServer) SemanticTokensFull(ctx context.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	return &protocol.SemanticTokens{}, nil
}

func (s *BaseServer) SemanticTokensFullDelta(ctx context.Context, params *protocol.SemanticTokensDeltaParams) (interface{}, error) {
	return &protocol.SemanticTokens{}, nil
}

func (s *BaseServer) SemanticTokensRange(ctx context.Context, params *protocol.SemanticTokensRangeParams) (*protocol.SemanticTokens, error) {
	return &protocol.SemanticTokens{}, nil
}

func (s *BaseServer) SemanticTokensRefresh(ctx context.Context) error {
	return nil
}

func (s *BaseServer) Moniker(ctx context.Context, params *protocol.MonikerParams) ([]protocol.Moniker, error) {
	return []protocol.Moniker{}, nil
}
