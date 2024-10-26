package provider

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/festeh/llm_flow/lsp/splitter"
)

type Dummy struct{}

var _ Provider = Dummy{} // Verify Dummy implements Provider interface

func (b *Dummy) Predict(ctx context.Context, w io.Writer, splitter splitter.SplitFn) error {
	data := make(map[string]string)
	if err := splitter(&data); err != nil {
		log.Println("error (!) splitting text into prefix and suffix")
		return err
	}
	log.Println("Being called with: ", data)
	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
			if _, err := fmt.Fprintln(w, data); err != nil {
				return err
			}
		}
	}
	return nil
}
