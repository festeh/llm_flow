package backend

import (
	"context"
	"io"
	"log"

	"github.com/festeh/llm_flow/lsp/provider"
)

type FimBackend struct {
	providers map[string]provider.Provider
}

func NewFimBackend() *FimBackend {
	providers := make(map[string]provider.Provider)
	for _, name := range []string{"codestral"} {
		if provider, err := provider.NewProvider(name); err == nil {
			log.Printf("Loaded provider: %s", name)
			providers[name] = provider
		}
	}
	return &FimBackend{}
}

func (b *FimBackend) Predict(ctx context.Context, w io.Writer, text string) error {
	return nil
}
