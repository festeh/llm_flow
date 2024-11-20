package main

import (
	"context"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/festeh/llm_flow/lsp"
)

func main() {
	log.SetTimeFormat(time.StampMilli)

	ctx := context.Background()
	server := lsp.NewServer(os.Stdout)

	if err := server.Serve(ctx, ":7777"); err != nil {
		log.Error("Server error: %v", err)
	}
}
