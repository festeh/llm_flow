package main

import (
	"context"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"log"
	"net"
	"os"
)

func main() {
	listener, err := net.Listen("tcp", ":7777")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Println("Server listening on :7777")
	
	server := &Server{}
	ctx := context.Background()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go func() {
			stream := jsonrpc2.NewStream(conn)
			rpcConn := jsonrpc2.NewConn(stream)
			handler := protocol.ServerHandler(server, jsonrpc2.MethodNotFoundHandler)
			rpcConn.Go(ctx, handler)
			<-rpcConn.Done()
		}()
	}
}
