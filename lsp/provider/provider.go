package provider

import (
	"fmt"

	"github.com/festeh/llm_flow/lsp/splitter"
)

type Response interface {
	Validate() error
	GetResult() string
}

type Provider interface {
	Name() string
	GetRequestBody(splitter.ProjectContext) (map[string]interface{}, error)
	GetAuthHeader() string
	Endpoint() string
	SetModel(string)
	IsStreaming() bool
	NewResponse() Response
}

func NewProvider(name string, model string) (Provider, error) {
	switch name {
	case "codestral":
		return newCodestral()
	case "huggingface":
		return newHuggingface(model)
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}
