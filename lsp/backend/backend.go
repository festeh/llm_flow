package backend

import (
	"context"
	"io"
)

// Backend defines the interface for prediction backends
type Backend interface {
	Predict(ctx context.Context, w io.Writer, text string, providerAndModel string) error
}
