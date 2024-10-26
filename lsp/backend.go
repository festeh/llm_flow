package lsp

import (
	"context"
	"io"
)

// Backend defines the interface for prediction backends
type Backend interface {
	Predict(ctx context.Context, w io.Writer, text string) error
}

// DummyBackend implements a simple dummy backend that just echoes text
type DummyBackend struct{}

func NewDummyBackend() *DummyBackend {
	return &DummyBackend{}
}

func (b *DummyBackend) Predict(ctx context.Context, w io.Writer, text string) error {
	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
			if _, err := fmt.Fprintln(w, text); err != nil {
				return err
			}
		}
	}
	return nil
}
