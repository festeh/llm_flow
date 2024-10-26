package provider

import (
	"fmt"
	"os"
)

type Provider interface {
}

func NewProvider(name string) (Provider, error) {
	switch name {
	case "codestral":
		return newCodestral()
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}

type Codestral struct {
	key string
}

func newCodestral() (*Codestral, error) {
	// get CODESTRAL_API_KEY from env
	key := os.Getenv("CODESTRAL_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("CODESTRAL_API_KEY not found")
	}
	return &Codestral{key: key}, nil
}
