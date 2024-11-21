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
}

func NewProvider(name string) (Provider, error) {
	switch name {
	case "codestral":
		return newCodestral()
	case "dummy":
		return Dummy{}, nil
  case "huggingface":
    return newHuggingface()
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}
