package lsp

import (
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/daulet/tokenizers"
	"github.com/festeh/llm_flow/lsp/provider"
)

type SetConfigParams struct {
	Repo     string `json:"repo"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

// Config holds server configuration
type Config struct {
	Repo      string
	Provider  *provider.Provider
	Tokenizer *tokenizers.Tokenizer
	Model     *string
}

func (c *Config) HandleSetConfig(params json.RawMessage) error {
	var configParams SetConfigParams
	if err := json.Unmarshal(params, &configParams); err != nil {
		return fmt.Errorf("error parsing set_config params: %v", err)
	}
	c.Repo = configParams.Repo
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
	c.Model = &model
	return nil
}

func (c *Config) SetTokenizer() error {
	if c.Provider == nil {
		return fmt.Errorf("provider not set")
	}

	tokenizer, err := tokenizers.FromPretrained(*c.Model)
	if err != nil {
		return fmt.Errorf("failed to create tokenizer: %v", err)
	}

	c.Tokenizer = tokenizer
	log.Info("Tokenizer initiazlied")
	return nil
}
