package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/festeh/llm_flow/lsp"
)

func main() {
	ctx := context.Background()
	server := lsp.NewServer()

	// Set up logging to a file
	logFile, err := os.OpenFile("lsp-server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// Main message processing loop
	for {
		message, err := readMessage(os.Stdin)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Error reading message: %v", err)
			continue
		}

		// Parse the JSON-RPC message
		var header struct {
			Method string          `json:"method"`
			ID     interface{}     `json:"id,omitempty"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.Unmarshal(message, &header); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Handle different methods
		var result interface{}
		var handleErr error

		switch header.Method {
		case "initialize":
			var params lsp.InitializeParams
			json.Unmarshal(header.Params, &params)
			result, handleErr = server.Initialize(ctx, &params)

		case "initialized":
			handleErr = server.Initialized(ctx)

		case "shutdown":
			handleErr = server.Shutdown(ctx)

		case "exit":
			handleErr = server.Exit(ctx)
			os.Exit(0)

		case "textDocument/didOpen":
			var params lsp.DidOpenTextDocumentParams
			json.Unmarshal(header.Params, &params)
			handleErr = server.TextDocumentDidOpen(ctx, &params)

		case "textDocument/didChange":
			var params lsp.DidChangeTextDocumentParams
			json.Unmarshal(header.Params, &params)
			handleErr = server.TextDocumentDidChange(ctx, &params)

		case "textDocument/completion":
			result, handleErr = server.TextDocumentCompletion(ctx, header.Params)

		default:
			log.Printf("Unknown method: %s", header.Method)
			continue
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
			sendResponse(response)
		}
	}
}

func readMessage(r io.Reader) ([]byte, error) {
	// Read headers
	var contentLength int
	for {
		header, err := readLine(r)
		if err != nil {
			return nil, err
		}
		if header == "" {
			break
		}
		if strings.HasPrefix(header, "Content-Length: ") {
			fmt.Sscanf(header, "Content-Length: %d", &contentLength)
		}
	}

	// Read content
	content := make([]byte, contentLength)
	_, err := io.ReadFull(r, content)
	return content, err
}

func readLine(r io.Reader) (string, error) {
	var line strings.Builder
	buf := make([]byte, 1)
	for {
		_, err := r.Read(buf)
		if err != nil {
			return "", err
		}
		if buf[0] == '\n' {
			return strings.TrimRight(line.String(), "\r"), nil
		}
		line.Write(buf)
	}
}

func sendResponse(response interface{}) {
	responseBytes, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return
	}

	fmt.Printf("Content-Length: %d\r\n", len(responseBytes))
	fmt.Printf("\r\n")
	fmt.Printf("%s", responseBytes)
}
