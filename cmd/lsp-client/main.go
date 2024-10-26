package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

type jsonrpcMessage struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

func main() {
	conn, err := net.Dial("tcp", "localhost:7777")
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Send predict request
	request := jsonrpcMessage{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "predict",
		Params:  map[string]string{"text": "sample text"},
	}

	// Marshal and send the request
	requestBytes, err := json.Marshal(request)
	if err != nil {
		log.Fatalf("Failed to marshal request: %v", err)
	}

	fmt.Fprintf(conn, "Content-Length: %d\r\n\r\n%s", len(requestBytes), requestBytes)

	// Read and process responses
	reader := bufio.NewReader(conn)
	for {
		message, err := readMessage(reader)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading message: %v", err)
			}
			return
		}

		var response jsonrpcMessage
		if err := json.Unmarshal(message, &response); err != nil {
			log.Printf("Error parsing response: %v", err)
			continue
		}

		if response.Method == "predict/response" {
			var predictResponse struct {
				Content string `json:"content"`
			}
			responseBytes, _ := json.Marshal(response.Params)
			json.Unmarshal(responseBytes, &predictResponse)
			fmt.Println("Received prediction:", predictResponse.Content)
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
			fmt.Sscanf(header, "Content-Length: %d", &contentLength)
		}
	}

	// Read content
	content := make([]byte, contentLength)
	_, err := io.ReadFull(r, content)
	return content, err
}
