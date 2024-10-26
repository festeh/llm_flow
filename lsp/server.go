package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/festeh/llm_flow/lsp/provider"
	"github.com/festeh/llm_flow/lsp/splitter"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

// Server represents an LSP server instance
type Server struct {
	documents map[string]string
	writer    io.Writer
	mu        sync.Mutex
	clients   map[net.Conn]struct{}
	providers map[string]provider.Provider
}

// NewServer creates a new LSP server instance
func NewServer(w io.Writer) *Server {
	providers := make(map[string]provider.Provider)
	for _, name := range []string{"codestral", "dummy"} {
		if provider, err := provider.NewProvider(name); err == nil {
			log.Printf("Loaded provider: %s", name)
			providers[name] = provider
		} else {
			log.Printf("Failed to load provider: %s", name)
		}
	}
	return &Server{
		documents: make(map[string]string),
		writer:    w,
		clients:   make(map[net.Conn]struct{}),
		providers: providers,
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
		var params struct {
			Text             string `json:"text"`
			ProviderAndModel string `json:"providerAndModel"`
		}
		if err := json.Unmarshal(header.Params, &params); err != nil {
			return fmt.Errorf("invalid predict params: %v", err)
		}
		if params.ProviderAndModel == "" {
			params.ProviderAndModel = "codestral/codestral-latest"
		}

		pr, pw := io.Pipe()
		go func() {
			defer pw.Close()
			if err := s.Predict(ctx, pw, params.Text, params.ProviderAndModel); err != nil {
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

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(responseBytes))
	message := append([]byte(header), responseBytes...)

	// Write the complete message atomically
	if _, err := s.writer.Write(message); err != nil {
		return fmt.Errorf("error writing response: %v", err)
	}

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
	// Add client to tracking
	s.mu.Lock()
	s.clients[conn] = struct{}{}
	s.mu.Unlock()

	// Ensure cleanup on exit
	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
		conn.Close()
		log.Printf("Client disconnected: %v", conn.RemoteAddr())
	}()

	log.Printf("New client connected: %v", conn.RemoteAddr())

	// Create a connection-specific writer
	connWriter := conn
	reader := bufio.NewReader(conn)

	// Create a connection-specific context that we can cancel
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Handle connection errors in a separate goroutine
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	for {
		message, err := readMessage(reader)
		if err != nil {
			if err == io.EOF {
				log.Printf("Client closed connection: %v", conn.RemoteAddr())
				return
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("Connection timeout: %v", conn.RemoteAddr())
				return
			}
			if strings.Contains(err.Error(), "connection reset by peer") {
				log.Printf("Client disconnected: %v", conn.RemoteAddr())
				return
			}
			log.Printf("Error reading message from %v: %v", conn.RemoteAddr(), err)
			return
		}

		// Set the writer for this specific connection
		s.mu.Lock()
		s.writer = connWriter
		s.mu.Unlock()

		if err := s.HandleMessage(ctx, message); err != nil {
			log.Printf("Error handling message from %v: %v", conn.RemoteAddr(), err)
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

func (s *Server) Predict(ctx context.Context, w io.Writer, text string, providerAndModel string) error {
	parts := strings.Split(providerAndModel, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid provider/model format: must be in format provider/model")
	}
	providerName := parts[0]
	model := parts[1]
	provider, ok := s.providers[providerName]
	if !ok {
		log.Println("provider not found")
		log.Printf("available providers: %d", len(s.providers))
		for k, v := range s.providers {
			log.Printf("%s: %s", k, v)
		}
		return fmt.Errorf("unknown provider: %s", providerName)
	}
	splitName := splitter.New(model, nil)
	var splitFn splitter.SplitFn
	switch splitName {
	case splitter.FimNaive:
		log.Println("FIM naive")
		splitFn = splitter.GetFimNaiveSplitter(text)
	default:
		return fmt.Errorf("unsupported splitter: %d", splitName)
	}
  log.Println("Predicting with: ", provider)
	return Flow(provider, splitFn, ctx, w)
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
