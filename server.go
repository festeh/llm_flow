package main

import (
	"context"
	"io"
	"time"

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
    return nil, nil
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

// Server embeds BaseServer to inherit all default implementations
type Server struct {
    BaseServer
}

func (s *Server) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: &protocol.TextDocumentSyncOptions{
				Change:    protocol.TextDocumentSyncKindFull,
				OpenClose: true,
			},
			CompletionProvider: &protocol.CompletionOptions{
				TriggerCharacters: []string{"."},
			},
		},
	}, nil
}

func (s *Server) Initialized(ctx context.Context, params *protocol.InitializedParams) error {
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}

func (s *Server) Exit(ctx context.Context) error {
	return nil
}

func (s *Server) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	// Handle document open
	return nil
}

func (s *Server) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	// Handle document changes
	return nil
}

func (s *Server) Completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	// Return dummy completion items
	return &protocol.CompletionList{
		IsIncomplete: false,
		Items: []protocol.CompletionItem{
			{
				Label: "Example",
				Kind:  protocol.CompletionItemKindText,
			},
		},
	}, nil
}

func (s *Server) CodeAction(ctx context.Context, params *protocol.CodeActionParams) ([]protocol.CodeAction, error) {
	return nil, nil
}

func (s *Server) Predict(ctx context.Context, writer io.Writer) error {
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	count := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if count >= 10 {
				return nil
			}
			_, err := writer.Write([]byte("foo"))
			if err != nil {
				return err
			}
			count++
		}
	}
}
