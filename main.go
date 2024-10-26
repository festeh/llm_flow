package main

import (
	"context"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"os"
)

func main() {
	stream := jsonrpc2.NewStream(os.Stdin)
	server := &Server{}
	ctx := context.Background()
	conn := jsonrpc2.NewConn(stream)
	handler := protocol.ServerHandler(server, jsonrpc2.MethodNotFoundHandler)
	conn.Go(ctx, handler)

	<-conn.Done()
}
