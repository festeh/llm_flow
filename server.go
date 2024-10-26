package main

import (
	"context"
	"io"
	"time"

	"go.lsp.dev/protocol"
)


// Server embeds BaseServer to inherit all default implementations
type Server struct {
    BaseServer
}

func (s *Server) ColorPresentation(ctx context.Context, params *protocol.ColorPresentationParams) ([]protocol.ColorPresentation, error) {
    return []protocol.ColorPresentation{
        {
            Label: "#FF0000",
            TextEdit: &protocol.TextEdit{
                Range: protocol.Range{
                    Start: protocol.Position{Line: 0, Character: 0},
                    End:   protocol.Position{Line: 0, Character: 7},
                },
                NewText: "#FF0000",
            },
        },
    }, nil
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
