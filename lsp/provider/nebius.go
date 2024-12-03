package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/festeh/llm_flow/lsp/splitter"
)

type Nebius struct {
	key       string
	model     string
	streaming bool
}

type NebiusResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

func (n *NebiusResponse) Validate() error {
	if len(n.Choices) == 0 {
		return fmt.Errorf("Got empty response")
	}
	return nil
}

func (n *NebiusResponse) GetResult() string {
	return n.Choices[0].Text
}

func (n *Nebius) Name() string {
	return "Nebius"
}

func (n *Nebius) Streaming() bool {
	return n.streaming
}

func newNebius(model string) (*Nebius, error) {
	key := os.Getenv("NEBIUS_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("NEBIUS_API_KEY not found")
	}
	return &Nebius{key: key, model: model}, nil
}

func (n *Nebius) GetRequestBody(ctx splitter.ProjectContext) (map[string]interface{}, error) {
	repoName := "<|repo_name|>"
	fileSep := "<|file_sep|>"
	fimPrefix := "<|fim_prefix|>"
	fimSuffix := "<|fim_suffix|>"
	fimMiddle := "<|fim_middle|>"
	repoBaseName := filepath.Base(ctx.Repo)
	relativeFilePath := strings.TrimPrefix(ctx.File, ctx.Repo+"/")
	prompt := fmt.Sprintf("%s%s\n%s%s\n%s%s%s%s%s",
		repoName, repoBaseName,
		fileSep, relativeFilePath,
		fimPrefix, ctx.Prefix,
		fimSuffix, ctx.Suffix, fimMiddle)
	fmt.Println(prompt)
	data := map[string]interface{}{
		"max_tokens":  32,
		"stream":      n.streaming,
		"model":       n.model,
		"temperature": 0,
		"prompt":      prompt,
		"stop":        []string{"<|file_sep|>"},
	}
	return data, nil
}

func (n *Nebius) GetAuthHeader() string {
	return "Bearer " + n.key
}

func (n *Nebius) Endpoint() string {
	return "https://api.studio.nebius.ai/v1/completions"
}

func (n *Nebius) SetModel(model string) {
	n.model = model
}

func (n *Nebius) IsStreaming() bool {
	return n.streaming
}

func (n *Nebius) NewResponse() Response {
	return &NebiusResponse{}
}
