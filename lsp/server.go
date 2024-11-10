package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/festeh/llm_flow/lsp/constants"
	"github.com/festeh/llm_flow/lsp/provider"
	"github.com/festeh/llm_flow/lsp/splitter"
	"io"
	"net"
	"strings"
	"sync"
)

// Server represents an LSP server instance
type Server struct {
	documents         map[string]string
	writer            io.Writer
	mu                sync.Mutex
	clients           map[net.Conn]struct{}
	providers         map[string]provider.Provider
	activePredictions map[int]context.CancelFunc
	predictionsMu     sync.Mutex
}

// NewServer creates a new LSP server instance
func NewServer(w io.Writer) *Server {
	providers := make(map[string]provider.Provider)
	for _, name := range []string{"codestral", "dummy"} {
		if provider, err := provider.NewProvider(name); err == nil {
			log.Info("Loaded", "provider", name)
			providers[name] = provider
		} else {
			log.Error("Failed to load provider: %s", name)
		}
	}
	return &Server{
		documents:         make(map[string]string),
		writer:            w,
		clients:           make(map[net.Conn]struct{}),
		providers:         providers,
		activePredictions: make(map[int]context.CancelFunc),
	}
}

type PredictEditorParams struct {
	ProviderAndModel string `json:"providerAndModel"`
	URI              string `json:"uri"`
	Line             int    `json:"line"`
	Pos              int    `json:"pos"`
}

// HandleMessage processes a single LSP message
func (s *Server) HandleMessage(ctx context.Context, message []byte) error {
	// Parse the JSON-RPC message
	var header struct {
		Method string          `json:"method"`
		ID     int             `json:"id,omitempty"`
		Params json.RawMessage `json:"params"`
	}
	header.ID = -1
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

	case "textDocument/didOpen":
		var params DidOpenTextDocumentParams
		json.Unmarshal(header.Params, &params)
		handleErr = s.TextDocumentDidOpen(ctx, &params)

	case "textDocument/didChange":
		var params DidChangeTextDocumentParams
		json.Unmarshal(header.Params, &params)
		handleErr = s.TextDocumentDidChange(ctx, &params)

	case "textDocument/didSave":
		var params DidSaveTextDocumentParams
		// log.Info("", "rawParams", string(header.Params))
		json.Unmarshal(header.Params, &params)
		handleErr = s.TextDocumentDidSave(ctx, &params)

	case "textDocument/completion":
		result, handleErr = s.TextDocumentCompletion(ctx, header.Params)

	case "cancel_predict_editor":
		s.predictionsMu.Lock()
		params := CancelParams{}
		err := json.Unmarshal(header.Params, &params)
		if err != nil {
			log.Info("Err in cancel", err)
		}
		id := params.ID
		log.Info("Cancel", "id", id)
		cancel, ok := s.activePredictions[id]
		if ok {
			cancel()
			delete(s.activePredictions, id)
			log.Info("Cancelled prediction %v", id)
		}
		s.predictionsMu.Unlock()
		return nil

	case "predict_editor":
		var params PredictEditorParams
		if err := json.Unmarshal(header.Params, &params); err != nil {
			return fmt.Errorf("invalid predict params: %v", err)
		}
		log.Debug("Got", "params", params)
		if params.ProviderAndModel == "" {
			params.ProviderAndModel = "codestral/codestral-latest"
		}

		// Create cancellable context
		predCtx, cancel := context.WithCancel(ctx)
		s.predictionsMu.Lock()
		s.activePredictions[header.ID] = cancel
		s.predictionsMu.Unlock()

		pr, pw := io.Pipe()
		go func() {
			defer pw.Close()
			content, err := s.PredictEditor(predCtx, pw, params)

			// Clean up prediction tracking
			s.predictionsMu.Lock()
			delete(s.activePredictions, header.ID)
			s.predictionsMu.Unlock()
			if err != nil {
				log.Error("Prediction", "error", err)
				return
			}
			// Send completion notification after prediction is done
			response := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      header.ID,
				"result": PredictResponse{
					ID:      header.ID,
					Content: content,
				},
			}
			s.sendResponse(response)
		}()

		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			// _ := map[string]interface{}{
			// 	"jsonrpc": "2.0",
			// 	"method":  "predict/response",
			// 	"params": PredictResponse{
			// 		Content: scanner.Text(),
			// 	},
			// }
			// s.sendResponse(response)
		}
		return nil

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
	if header.ID != -1 {
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
		marsh, _ := json.Marshal(response)
		log.Info("Sending", "resp", string(marsh))
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

	log.Info("LSP server listening on", "port", addr)

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
		Info: ServerInfo{
			Name:    "llm_flow",
			Version: "0.0.1",
		},
		Capabilities: ServerCapabilities{
			TextDocumentSync: TextDocumentSyncOptions{
				OpenClose: true,
				Change:    1, // Full document sync
				Save: SaveOptions{
					IncludeText: true,
				},
			},
			CompletionProvider: false,
		},
	}, nil
}

// Initialized handles the LSP initialized notification
func (s *Server) Initialized(ctx context.Context) error {
	log.Info("Server initialized")
	return nil
}

// Shutdown handles the LSP shutdown request
func (s *Server) Shutdown(ctx context.Context) error {
	log.Info("Shutdown request received")
	return nil
}

// Exit handles the LSP exit notification
func (s *Server) Exit(ctx context.Context) error {
	log.Info("Exit notification received")
	return nil
}

// TextDocumentDidOpen handles textDocument/didOpen notification
func (s *Server) TextDocumentDidOpen(ctx context.Context, params *DidOpenTextDocumentParams) error {
	log.Info("Opened:", "uri", params.TextDocument.URI)
	s.documents[params.TextDocument.URI] = params.TextDocument.Text
	return nil
}

// TextDocumentDidChange handles textDocument/didChange notification
func (s *Server) TextDocumentDidChange(ctx context.Context, params *DidChangeTextDocumentParams) error {
	log.Info("Changed:", "uri", params.TextDocument.URI, "len",
		len(params.ContentChanges[0].Text))
	// For now, just store the full content
	if len(params.ContentChanges) > 0 {
		s.documents[params.TextDocument.URI] = params.ContentChanges[0].Text
	}
	return nil
}

// TextDocumentDidSave handles textDocument/didSave notification
func (s *Server) TextDocumentDidSave(ctx context.Context, params *DidSaveTextDocumentParams) error {
	text := params.TextDocument.Text
	log.Info("Saved:", "uri", params.TextDocument.URI, "len", len(text))
	if len(text) > 0 {
		s.documents[params.TextDocument.URI] = text
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
		log.Error("provider not found")
		log.Error("available providers: ")
		for k, v := range s.providers {
			log.Error("%s: %s", k, v)
		}
		return fmt.Errorf("unknown provider: %s", providerName)
	}
	splitName := splitter.New(model, nil)
	var splitFn splitter.SplitFn
	switch splitName {
	case splitter.FimNaive:
		splitFn = splitter.GetFimNaiveSplitter(text)
	default:
		return fmt.Errorf("unsupported splitter: %d", splitName)
	}
	_, err := Flow(provider, splitFn, ctx, w)
	return err
}

func (s *Server) PredictEditor(ctx context.Context, w io.Writer, params PredictEditorParams) (string, error) {
	parts := strings.Split(params.ProviderAndModel, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid provider/model format: must be in format provider/model")
	}
	providerName := parts[0]
	model := parts[1]
	provider, ok := s.providers[providerName]
	if !ok {
		log.Error("provider not found")
		for k, v := range s.providers {
			log.Error("%s: %s", k, v)
		}
		return "", fmt.Errorf("unknown provider: %s", providerName)
	}
	splitName := splitter.New(model, nil)
	var splitFn splitter.SplitFn
	switch splitName {
	case splitter.FimNaive:
		// Get document content
		doc, exists := s.documents[params.URI]
		if !exists {
			return "", fmt.Errorf("document not found: %s", params.URI)
		}

		// Split document into lines
		lines := strings.Split(doc, "\n")
		if params.Line >= len(lines) {
			return "", fmt.Errorf("line number out of range: %d", params.Line)
		}

		currentLine := lines[params.Line]
		prefix := strings.Join(lines[:params.Line], "\n")
		if params.Line > 0 {
			prefix += "\n"
		}

		pos := params.Pos
		suffix := ""
		if pos >= len(currentLine) {
			prefix += currentLine
		} else {
			prefix += currentLine[:pos]
			suffix = currentLine[pos:]
		}

		if params.Line < len(lines)-1 {
			suffix += "\n" + strings.Join(lines[params.Line+1:], "\n")
		}

		// Combine with FIM token
		text := prefix + constants.FIM_TOKEN + suffix
		splitFn = splitter.GetFimNaiveSplitter(text)
	default:
		return "", fmt.Errorf("unsupported splitter: %d", splitName)
	}
	return Flow(provider, splitFn, ctx, w)
}

type DidSaveTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type CancelParams struct {
	ID int `json:"id"`
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
