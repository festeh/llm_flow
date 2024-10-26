package lsp

import (
	"context"
	"io"
	"log"

	"github.com/festeh/llm_flow/lsp/provider"
	"github.com/festeh/llm_flow/lsp/splitter"
)

func Flow(p provider.Provider, splitter splitter.SplitFn, ctx context.Context, w io.Writer) error {
	log.Printf("Flow: using provider %s", p.Name())
	return p.Predict(ctx, w, splitter)
}
