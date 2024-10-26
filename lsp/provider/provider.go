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
	Predict(ctx context.Context, w io.Writer, splitter splitter.SplitFn) error
	Name() string
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

func (c *Codestral) Predict(ctx context.Context, w io.Writer, splitter splitter.SplitFn) error {
	data := make(map[string]string)
	if err := splitter(&data); err != nil {
		return fmt.Errorf("error splitting text: %v", err)
	}

	reqBody := map[string]interface{}{
		"model":       "codestral-latest",
		"prompt":      data["prefix"],
		"suffix":      data["suffix"],
		"max_tokens":  64,
		"temperature": 0,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.mistral.ai/v1/fim/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.key)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var streamResp struct {
			Choices []struct {
				Text string `json:"text"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &streamResp); err != nil {
			return fmt.Errorf("error parsing response: %v", err)
		}
		if len(streamResp.Choices) > 0 {
			if _, err := fmt.Fprintln(w, streamResp.Choices[0].Text); err != nil {
				return fmt.Errorf("error writing response: %v", err)
			}
		}
	}

	return scanner.Err()
}
