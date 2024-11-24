package lsp

import "github.com/festeh/llm_flow/lsp/provider"

// Config holds server configuration
type Config struct {
	Provider *provider.Provider
}

func (c *Config) SetProvider(providerName string, model string) error {
	p, err := provider.NewProvider(providerName, model)
	if err != nil {
		return err
	}
	c.Provider = &p
	return nil
}
