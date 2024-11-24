package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/festeh/llm_flow/lsp"
)

func main() {
	port := flag.Int("port", 7777, "Server port to listen on")
	flag.Parse()

	log.SetTimeFormat(time.StampMilli)

	ctx := context.Background()
	server := lsp.NewServer(os.Stdout)

	addr := fmt.Sprintf(":%d", *port)
	if err := server.Serve(ctx, addr); err != nil {
		log.Error("Server error: %v", err)
	}
}
