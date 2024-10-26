package lsp

import (
	"context"
	"encoding/json"
	"log"
)

// Server represents an LSP server instance
type Server struct {
	documents map[string]string
}

// NewServer creates a new LSP server instance
func NewServer() *Server {
	return &Server{
		documents: make(map[string]string),
	}
}

// Initialize handles the LSP initialize request
func (s *Server) Initialize(ctx context.Context, params *InitializeParams) (*InitializeResult, error) {
	log.Printf("Initialize request received. Root URI: %s", params.RootURI)
	
	return &InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync:    1, // Incremental sync
			CompletionProvider: true,
		},
	}, nil
}

// Initialized handles the LSP initialized notification
func (s *Server) Initialized(ctx context.Context) error {
	log.Println("Server initialized")
	return nil
}

// Shutdown handles the LSP shutdown request
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutdown request received")
	return nil
}

// Exit handles the LSP exit notification
func (s *Server) Exit(ctx context.Context) error {
	log.Println("Exit notification received")
	return nil
}

// TextDocumentDidOpen handles textDocument/didOpen notification
func (s *Server) TextDocumentDidOpen(ctx context.Context, params *DidOpenTextDocumentParams) error {
	log.Printf("Document opened: %s", params.TextDocument.URI)
	s.documents[params.TextDocument.URI] = params.TextDocument.Text
	return nil
}

// TextDocumentDidChange handles textDocument/didChange notification
func (s *Server) TextDocumentDidChange(ctx context.Context, params *DidChangeTextDocumentParams) error {
	log.Printf("Document changed: %s", params.TextDocument.URI)
	// For now, just store the full content
	if len(params.ContentChanges) > 0 {
		s.documents[params.TextDocument.URI] = params.ContentChanges[0].Text
	}
	return nil
}

// TextDocumentCompletion handles textDocument/completion request
func (s *Server) TextDocumentCompletion(ctx context.Context, params json.RawMessage) (*CompletionList, error) {
	// Dummy implementation returning some static completion items
	return &CompletionList{
		IsIncomplete: false,
		Items: []CompletionItem{
			{
				Label:  "example",
				Kind:   1, // Text
				Detail: "Example completion item",
			},
		},
	}, nil
}
