package provider

import (
	"context"
	"fmt"
	"io"
	"os"
)

type Provider interface {
	// Name() string
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

func (c *Codestral) Predict(ctx context.Context, w io.Writer, text string, model string) error {
	return nil
}
