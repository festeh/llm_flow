package lsp

import (
	"encoding/json"
	"fmt"
	"github.com/daulet/tokenizers"
	"github.com/festeh/llm_flow/lsp/provider"
)

type SetConfigParams struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

// Config holds server configuration
type Config struct {
	Provider  *provider.Provider
	Tokenizer *tokenizers.Tokenizer
}

func (c *Config) HandleSetConfig(params json.RawMessage) error {
	var configParams SetConfigParams
	if err := json.Unmarshal(params, &configParams); err != nil {
		return fmt.Errorf("error parsing set_config params: %v", err)
	}
	if err := c.SetProvider(configParams.Provider, configParams.Model); err != nil {
		return err
	}
	return c.SetTokenizer()
}

func (c *Config) SetProvider(providerName string, model string) error {
	p, err := provider.NewProvider(providerName, model)
	if err != nil {
		return err
	}
	c.Provider = &p
	return nil
}

func (c *Config) SetTokenizer() error {
	if c.Provider == nil {
		return fmt.Errorf("provider not set")
	}
	
	tokenizer, err := tokenizers.FromPretrained((*c.Provider).Name())
	if err != nil {
		return fmt.Errorf("failed to create tokenizer: %v", err)
	}
	
	c.Tokenizer = &tokenizer
	return nil
}
