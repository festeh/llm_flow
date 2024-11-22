package lsp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/log"
	"io"
	"net/http"
	"strings"

	"github.com/festeh/llm_flow/lsp/provider"
	"github.com/festeh/llm_flow/lsp/splitter"
)

func Flow(p provider.Provider, prefixSuffix splitter.PrefixSuffix, ctx context.Context, w io.Writer) (string, error) {

	if p, ok := p.(*provider.Dummy); ok {
		err := p.Predict(ctx, w, prefixSuffix)
		return "dummy result", err
	}

	var buffer strings.Builder
	reqBody, err := p.GetRequestBody(prefixSuffix)
	if err != nil {
		return "", fmt.Errorf("error getting request body: %v", err)
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.Endpoint(), bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", p.GetAuthHeader())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()
	log.Info("request was sent")
	if p.IsStreaming() {
		if err := handleStreamingResponse(ctx, resp.Body, &buffer); err != nil {
			return "", err
		}
	} else {
		if err := handleNonStreamingResponse(resp.Body, &buffer); err != nil {
			return "", err
		}
	}


func handleStreamingResponse(ctx context.Context, body io.ReadCloser, buffer *strings.Builder) error {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			log.Info("Flow is cancelled")
			return ctx.Err()
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
				return fmt.Errorf("error parsing response: %v", err)
			}
			choice := streamResp.Choices[0].Delta.Content
			buffer.WriteString(choice)
		}
	}
	return scanner.Err()
}

func handleNonStreamingResponse(body io.ReadCloser, buffer *strings.Builder) error {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}
	var response struct {
		Choices []struct {
			Text string `json:"text"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return fmt.Errorf("error parsing response: %v", err)
	}
	if len(response.Choices) == 0 {
		return fmt.Errorf("no choices in response")
	}
	buffer.WriteString(response.Choices[0].Text)
	return nil
}

	res := buffer.String()
	log.Debug("Done", "result", res)
	return res, nil
}
