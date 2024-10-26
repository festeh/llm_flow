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

func (b *FimBackend) Predict(ctx context.Context, w io.Writer, text string, providerAndModel string) error {
	// Split text into prefix and suffix using FIM token
	parts := strings.Split(text, "<|FIM|>")
	if len(parts) != 2 {
		return fmt.Errorf("invalid text format: must contain exactly one <|FIM|> token")
	}
	prefix := parts[0]
	suffix := parts[1]

	// Split provider and model
	providerParts := strings.Split(providerAndModel, "/")
	if len(providerParts) != 2 {
		return fmt.Errorf("invalid provider/model format: must be in format provider/model")
	}
	providerName := providerParts[0]
	model := providerParts[1]

	// Get the provider
	provider, ok := b.providers[providerName]
	if !ok {
		return fmt.Errorf("unknown provider: %s", providerName)
	}

	// For now, we only support Codestral
	if providerName == "codestral" {
		if codestralProvider, ok := provider.(*provider.Codestral); ok {
			return codestralProvider.Predict(ctx, w, prefix, suffix, model)
		}
	}

	return fmt.Errorf("unsupported provider: %s", providerName)
}
