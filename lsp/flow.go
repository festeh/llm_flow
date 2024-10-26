package lsp

import (
	"context"
	"io"
	"log"

	"github.com/festeh/llm_flow/lsp/provider"
	"github.com/festeh/llm_flow/lsp/splitter"
)

func Flow(p provider.Provider, splitter splitter.SplitFn, ctx context.Context, w io.Writer) error {
	// check for DummyProvider
	if provider, ok := p.(provider.Dummy); ok {
		log.Println("Flow: dummy")
		return provider.Predict(ctx, w, splitter)
	}
	log.Println("Not a dummy provider")
	return nil
}
