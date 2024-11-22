package provider

import (
	"fmt"
	"os"

	"github.com/festeh/llm_flow/lsp/splitter"
)

type Huggingface struct {
	key       string
	model     string
	streaming bool
}

func (c *Huggingface) Name() string {
	return "Huggingface"
}

func (c *Huggingface) Streaming() bool {
	return c.streaming
}

func newHuggingface() (*Huggingface, error) {
	key := os.Getenv("HF_API_TOKEN")
	if key == "" {
		return nil, fmt.Errorf("HF_API_TOKEN not found")
	}
	return &Huggingface{key: key, model: "codellama/CodeLlama-13b-hf"}, nil
}

func (c *Huggingface) GetRequestBody(prefixSuffix splitter.PrefixSuffix) (map[string]interface{}, error) {
	parameters := map[string]interface{}{
		"max_tokens":  32,
		"temperature": 0,
		"stream":      c.streaming,
	}

	input := fmt.Sprintf("<PRE> %s <SUF>%s <MID>", prefixSuffix.Prefix, prefixSuffix.Suffix)

	data := map[string]interface{}{
		"parameters": parameters,
		"stream":     c.streaming,
		"inputs":     input,
	}

	return data, nil
}

func (c *Huggingface) GetAuthHeader() string {
	return "Bearer " + c.key
}

func (c *Huggingface) Endpoint() string {
	return "https://api-inference.huggingface.co/models/" + c.model
}

func (c *Huggingface) SetModel(model string) {
	c.model = model
}

func (c *Huggingface) IsStreaming() bool {
	return c.streaming
}
