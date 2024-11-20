package main

import (
	"context"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/festeh/llm_flow/lsp"
)

func main() {
	// Set up logging to a file
	// logFile, err := os.OpenFile("lsp-server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// if err != nil {
	// 	panic(err)
	// }
	// defer logFile.Close()
	// log.SetOutput(logFile)

  log.SetTimeFormat(time.StampMilli)

	ctx := context.Background()
	server := lsp.NewServer(os.Stdout)

	if err := server.Serve(ctx, ":7777"); err != nil {
		log.Error("Server error: %v", err)
	}
}
