package lsp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/festeh/llm_flow/lsp/provider"
	"github.com/festeh/llm_flow/lsp/splitter"
)

func Flow(p provider.Provider, splitter splitter.SplitFn, ctx context.Context, w io.Writer) error {
	log.Printf("Flow: using provider %s", p.Name())

	data := make(map[string]string)
	if err := splitter(&data); err != nil {
		return fmt.Errorf("error splitting text: %v", err)
	}

	reqBody, err := p.GetRequestBody(data["prefix"], data["suffix"])
	if err != nil {
		return fmt.Errorf("error getting request body: %v", err)
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
	req.Header.Set("Authorization", p.GetAuthHeader())

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
