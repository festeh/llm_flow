package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Provider interface {
	Name() string
	GetRequestBody(prefix, suffix string) (map[string]interface{}, error)
	GetAuthHeader() string
}

func NewProvider(name string) (Provider, error) {
	switch name {
	case "codestral":
		return newCodestral()
	case "dummy":
		return Dummy{}, nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}

type Codestral struct {
	key string
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
	return &Codestral{key: key}, nil
}

func (c *Codestral) GetRequestBody(prefix, suffix string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"model":       "codestral-latest",
		"prompt":      prefix,
		"suffix":      suffix,
		"max_tokens":  64,
		"temperature": 0,
	}, nil
}

func (c *Codestral) GetAuthHeader() string {
	return "Bearer " + c.key
}
