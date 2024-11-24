package lsp

import (
	"encoding/json"
	"fmt"
	"github.com/festeh/llm_flow/lsp/provider"
)

type SetConfigParams struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

// Config holds server configuration
type Config struct {
	Provider *provider.Provider
}

func (c *Config) HandleSetConfig(params json.RawMessage) error {
	var configParams SetConfigParams
	if err := json.Unmarshal(params, &configParams); err != nil {
		return fmt.Errorf("error parsing set_config params: %v", err)
	}
	return c.SetProvider(configParams.Provider, configParams.Model)
}

func (c *Config) SetProvider(providerName string, model string) error {
	p, err := provider.NewProvider(providerName, model)
	if err != nil {
		return err
	}
	c.Provider = &p
	return nil
}
