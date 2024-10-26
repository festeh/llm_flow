package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"go.lsp.dev/jsonrpc2"
)

type PredictClient struct {
	conn jsonrpc2.Conn
}

func (c *PredictClient) Predict(ctx context.Context) error {
	// Create a pipe to receive streaming results
	fmt.Println("Called Predict")
	pr, pw := io.Pipe()

	// Start goroutine to read from pipe and print to stdout
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := pr.Read(buf)
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Printf("Error reading prediction: %v", err)
				return
			}
			fmt.Print(string(buf[:n]))
		}
	}()

	// Call Predict method
	id, err := c.conn.Call(ctx, "Predict", nil, pw)
	fmt.Println(id)
	if err != nil {
		return fmt.Errorf("predict call failed: %w", err)
	}

	return nil
}

func main() {
	ctx := context.Background()

	// Connect to server on port 7777
	conn, err := net.Dial("tcp", "localhost:7777")
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	log.Println("Connected")
	defer conn.Close()

	stream := jsonrpc2.NewStream(conn)
	rpcConn := jsonrpc2.NewConn(stream)

	client := &PredictClient{conn: rpcConn}

	// Call predict and print results
	if err := client.Predict(ctx); err != nil {
		log.Fatalf("Prediction failed: %v", err)
	}

	// <-conn
}
