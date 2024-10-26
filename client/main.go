package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
)

type PredictClient struct {
	conn *jsonrpc2.Conn
}

func (c *PredictClient) Predict(ctx context.Context) error {
	// Create a pipe to receive streaming results
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
	err := c.conn.Call(ctx, "Predict", nil, pw)
	if err != nil {
		return fmt.Errorf("predict call failed: %w", err)
	}

	return nil
}

func main() {
	ctx := context.Background()

	// Connect to server's stdin/stdout
	stream := jsonrpc2.NewStream(os.Stdin)
	conn := jsonrpc2.NewConn(stream)
	
	client := &PredictClient{conn: conn}

	// Call predict and print results
	if err := client.Predict(ctx); err != nil {
		log.Fatalf("Prediction failed: %v", err)
	}

	<-conn.Done()
}
