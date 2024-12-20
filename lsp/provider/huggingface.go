package provider

import (
	"fmt"
	"os"

	"github.com/festeh/llm_flow/lsp/splitter"
)

type Huggingface struct {
	key       string
	model     string
	streaming bool
}

type HuggingfaceResponse []struct {
	GeneratedText string `json:"generated_text"`
}

func (h *HuggingfaceResponse) Validate() error {
	if len(*h) == 0 {
		return fmt.Errorf("no items in Huggingface response")
	}
	if (*h)[0].GeneratedText == "" {
		return fmt.Errorf("generated text is empty")
	}
	return nil
}

func (h *HuggingfaceResponse) GetResult() string {
	return (*h)[0].GeneratedText
}

func (c *Huggingface) Name() string {
	return "Huggingface"
}

func (c *Huggingface) Streaming() bool {
	return c.streaming
}

func newHuggingface(model string) (*Huggingface, error) {
	key := os.Getenv("HF_API_TOKEN")
	if key == "" {
		return nil, fmt.Errorf("HF_API_TOKEN not found")
	}
	return &Huggingface{key: key, model: model}, nil
}

func (c *Huggingface) GetRequestBody(ctx splitter.ProjectContext) (map[string]interface{}, error) {
	parameters := map[string]interface{}{
		"max_tokens":       32,
		"stream":           c.streaming,
		"return_full_text": false,
	}

	input := fmt.Sprintf("%s\n▁<PRE> %s ▁<SUF>%s ▁<MID>", ctx.File, ctx.Prefix, ctx.Suffix)

	data := map[string]interface{}{
		"parameters": parameters,
		"stream":     c.streaming,
		"inputs":     input,
	}

	return data, nil
}

func (c *Huggingface) GetAuthHeader() string {
	return "Bearer " + c.key
}

func (c *Huggingface) Endpoint() string {
	return "https://api-inference.huggingface.co/models/" + c.model
}

func (c *Huggingface) SetModel(model string) {
	c.model = model
}

func (c *Huggingface) IsStreaming() bool {
	return c.streaming
}

func (c *Huggingface) NewResponse() Response {
	return &HuggingfaceResponse{}
}
