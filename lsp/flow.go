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
	"strings"

	"github.com/festeh/llm_flow/lsp/provider"
	"github.com/festeh/llm_flow/lsp/splitter"
)

func Flow(p provider.Provider, splitter splitter.SplitFn, ctx context.Context, w io.Writer) error {

	if p, ok := p.(*provider.Dummy); ok {
		return p.Predict(ctx, w, splitter)
	}

	reqBody, err := p.GetRequestBody(splitter)
	if err != nil {
		return fmt.Errorf("error getting request body: %v", err)
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error marshaling request: %v", err)
	}

	fmt.Println(string(jsonBody))

	req, err := http.NewRequestWithContext(ctx, "POST", p.Endpoint(), bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", p.GetAuthHeader())

	client := &http.Client{}
	resp, err := client.Do(req)
	fmt.Println(resp)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		content := strings.TrimPrefix(line, "data: ")
		if content == "[DONE]" {
			break
		}
		log.Println(content)
		var streamResp struct {
			Choices []struct {
				Text string `json:"text"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(content), &streamResp); err != nil {
			return fmt.Errorf("error parsing response: %v", err)
		}
		choice := streamResp.Choices[0].Text
		if _, err := w.Write([]byte(choice)); err != nil {
			return fmt.Errorf("error writing response: %v", err)
		}
	}

	return scanner.Err()
}
