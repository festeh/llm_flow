package main

import (
	"context"
	"os"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
)

func main() {
	ctx := context.Background()

	stream := jsonrpc2.NewStream(jsonrpc2.NewStream(os.Stdin, os.Stdout))
	conn := jsonrpc2.NewConn(stream)
	
	server := &Server{}
	protocol.ServerHandler(server, conn, jsonrpc2.MethodNotFound)

	<-conn.Done()
}
