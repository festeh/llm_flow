package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

// Server represents an LSP server instance
type Server struct {
	documents map[string]string
	writer    io.Writer
}

// NewServer creates a new LSP server instance
func NewServer(w io.Writer) *Server {
	return &Server{
		documents: make(map[string]string),
		writer:    w,
	}
}

// HandleMessage processes a single LSP message
func (s *Server) HandleMessage(ctx context.Context, message []byte) error {
	// Parse the JSON-RPC message
	var header struct {
		Method string          `json:"method"`
		ID     interface{}     `json:"id,omitempty"`
		Params json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal(message, &header); err != nil {
		return fmt.Errorf("error parsing message: %v", err)
	}

	// Handle different methods
	var result interface{}
	var handleErr error

	switch header.Method {
	case "initialize":
		var params InitializeParams
		json.Unmarshal(header.Params, &params)
		result, handleErr = s.Initialize(ctx, &params)

	case "initialized":
		handleErr = s.Initialized(ctx)

	case "shutdown":
		handleErr = s.Shutdown(ctx)

	case "exit":
		handleErr = s.Exit(ctx)
		os.Exit(0)

	case "textDocument/didOpen":
		var params DidOpenTextDocumentParams
		json.Unmarshal(header.Params, &params)
		handleErr = s.TextDocumentDidOpen(ctx, &params)

	case "textDocument/didChange":
		var params DidChangeTextDocumentParams
		json.Unmarshal(header.Params, &params)
		handleErr = s.TextDocumentDidChange(ctx, &params)

	case "textDocument/completion":
		result, handleErr = s.TextDocumentCompletion(ctx, header.Params)

	case "predict":
		pr, pw := io.Pipe()
		go func() {
			defer pw.Close()
			if err := s.Predict(ctx, pw); err != nil {
				log.Printf("Prediction error: %v", err)
			}
			// Send completion notification after prediction is done
			response := map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "predict/complete",
			}
			s.sendResponse(response)
		}()

		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			response := map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "predict/response",
				"params": PredictResponse{
					Content: scanner.Text(),
				},
			}
			s.sendResponse(response)
		}
		return nil

	default:
		return fmt.Errorf("unknown method: %s", header.Method)
	}

	// Send response for requests (methods with IDs)
	if header.ID != nil {
		response := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      header.ID,
		}
		if handleErr != nil {
			response["error"] = map[string]interface{}{
				"code":    -32603,
				"message": handleErr.Error(),
			}
		} else {
			response["result"] = result
		}
		s.sendResponse(response)
	}

	return handleErr
}

func (s *Server) sendResponse(response interface{}) error {
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("error marshaling response: %v", err)
	}

	fmt.Fprintf(s.writer, "Content-Length: %d\r\n", len(responseBytes))
	fmt.Fprintf(s.writer, "\r\n")
	fmt.Fprintf(s.writer, "%s", responseBytes)
	return nil
}

// Serve starts the LSP server on the specified address
func (s *Server) Serve(ctx context.Context, addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}
	defer listener.Close()

	log.Printf("LSP server listening on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go s.handleConnection(ctx, conn)
	}
}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	defer func() {
		log.Printf("Client disconnected: %v", conn.RemoteAddr())
		conn.Close()
	}()

	log.Printf("New client connected: %v", conn.RemoteAddr())
	s.writer = conn
	reader := bufio.NewReader(conn)

	for {
		message, err := readMessage(reader)
		if err != nil {
			if err == io.EOF {
				return
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("Connection timeout: %v", err)
				return
			}
			if strings.Contains(err.Error(), "connection reset by peer") {
				log.Printf("Client disconnected unexpectedly: %v", err)
				return
			}
			log.Printf("Error reading message: %v", err)
			return
		}

		if err := s.HandleMessage(ctx, message); err != nil {
			log.Printf("Error handling message: %v", err)
		}
	}
}

func readMessage(r *bufio.Reader) ([]byte, error) {
	var contentLength int

	// Read headers
	for {
		header, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		header = strings.TrimSpace(header)
		if header == "" {
			break
		}
		if strings.HasPrefix(header, "Content-Length: ") {
			if _, err := fmt.Sscanf(header, "Content-Length: %d", &contentLength); err != nil {
				return nil, fmt.Errorf("invalid Content-Length header: %v", err)
			}
		}
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("no Content-Length header found")
	}

	// Read exactly contentLength bytes
	content := make([]byte, contentLength)
	if _, err := io.ReadFull(r, content); err != nil {
		return nil, fmt.Errorf("failed to read message content: %v", err)
	}

	// Verify it's valid JSON
	if !json.Valid(content) {
		return nil, fmt.Errorf("invalid JSON content: %s", content)
	}

	return content, nil
}

// Initialize handles the LSP initialize request
func (s *Server) Initialize(ctx context.Context, params *InitializeParams) (*InitializeResult, error) {
	log.Printf("Initialize request received. Root URI: %s", params.RootURI)

	return &InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync:   1, // Incremental sync
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

// Predict streams predictions with a delay between each one
func (s *Server) Predict(ctx context.Context, w io.Writer) error {
	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(30 * time.Millisecond):
			if _, err := fmt.Fprintln(w, "foo"); err != nil {
				return err
			}
		}
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
