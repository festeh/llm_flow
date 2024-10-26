package provider

import (
	"fmt"
	"os"

	"github.com/festeh/llm_flow/lsp/splitter"
)

type Codestral struct {
	key   string
	model string
}

func (c *Codestral) Name() string {
	return "codestral"
}

func newCodestral() (*Codestral, error) {
	// get CODESTRAL_API_KEY from env
	key := os.Getenv("CODESTRAL_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("CODESTRAL_API_KEY not found")
	}
	return &Codestral{key: key, model: "codestral-latest"}, nil
}

func (c *Codestral) GetRequestBody(splitFn splitter.SplitFn) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"model": c.model,
		"max_tokens":  64,
		"temperature": 0,
		"stream": true,
	}

	if err := splitFn(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func (c *Codestral) GetAuthHeader() string {
	return "Bearer " + c.key
}

func (c *Codestral) Endpoint() string {
	return "https://codestral.mistral.ai/v1/fim/completions"
}
