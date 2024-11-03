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

func Flow(p provider.Provider, splitter splitter.SplitFn, ctx context.Context, w io.Writer) (string, error) {

	if p, ok := p.(*provider.Dummy); ok {
		err := p.Predict(ctx, w, splitter)
		return "dummy result", err
	}

	var buffer strings.Builder
	reqBody, err := p.GetRequestBody(splitter)
	if err != nil {
		return "", fmt.Errorf("error getting request body: %v", err)
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	fmt.Println(string(jsonBody))

	req, err := http.NewRequestWithContext(ctx, "POST", p.Endpoint(), bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", p.GetAuthHeader())

	client := &http.Client{}
	resp, err := client.Do(req)
	log.Println(resp)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			log.Println("Cancelled")
			return "", ctx.Err()
		default:
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			content := strings.TrimPrefix(line, "data: ")
			if content == "[DONE]" {
				break
			}
			var streamResp struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(content), &streamResp); err != nil {
				return "", fmt.Errorf("error parsing response: %v", err)
			}
			choice := streamResp.Choices[0].Delta.Content
			buffer.WriteString(choice)
			if _, err = fmt.Fprint(w, choice); err != nil {
				return "", fmt.Errorf("error writing response: %v", err)
			}
		}
		if err := scanner.Err(); err != nil {
			return "", err
		}
	}
	res := buffer.String()
	log.Println("Result", res)
	return res, nil
}
