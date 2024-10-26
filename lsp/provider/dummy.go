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

func (b Dummy) Name() string {
	return "dummy"
}

func (b *Dummy) Predict(ctx context.Context, w io.Writer, splitter splitter.SplitFn) error {
	data := make(map[string]interface{})
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

func (b Dummy) GetAuthHeader() string {
	return "dummy"
}

func (b Dummy) GetRequestBody(splitter.SplitFn) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (b Dummy) Endpoint() string {
  return "dummy"
}
