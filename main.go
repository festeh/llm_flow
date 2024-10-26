package main

import (
	"context"
	"os"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
)

func main() {
	ctx := context.Background()

	stream := jsonrpc2.NewStream(os.Stdin)
	conn := jsonrpc2.NewConn(ctx, stream, nil)
	
	server := &Server{}
	handler := protocol.ServerHandler(server, conn)
	handler.Handle(ctx)

	<-conn.Done()
}
