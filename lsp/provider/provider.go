package provider

import (
	"fmt"

	"github.com/festeh/llm_flow/lsp/splitter"
)

type Provider interface {
	Name() string
	GetRequestBody(splitter.PrefixSuffix) (map[string]interface{}, error)
	GetAuthHeader() string
  Endpoint() string
  SetModel(string)
  IsStreaming() bool
}

func NewProvider(name string) (Provider, error) {
	switch name {
	case "codestral":
		return newCodestral()
  case "huggingface":
    return newHuggingface()
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}
