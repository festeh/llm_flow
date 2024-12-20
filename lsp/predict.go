package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/festeh/llm_flow/lsp/splitter"
	"io"
	"strings"
)

func (s *Server) HandlePredictRequest(ctx context.Context, params json.RawMessage, header Header) error {
	var predictParams struct {
		Text             string `json:"text"`
		ProviderAndModel string `json:"providerAndModel"`
	}
	if err := json.Unmarshal(params, &predictParams); err != nil {
		return fmt.Errorf("invalid predict params: %v", err)
	}
	if predictParams.ProviderAndModel == "" {
		predictParams.ProviderAndModel = "codestral/codestral-latest"
	}

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		if err := s.Predict(ctx, pw, predictParams.Text, predictParams.ProviderAndModel); err != nil {
			log.Printf("Prediction error: %v", err)
		}
		// Send completion notification after prediction is done
		response := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  "predict/complete",
		}
		s.sendResponse(response)
	}()

	scanner := bufio.NewScanner(pr)
	for scanner.Scan() {
		response := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  "predict/response",
			"params": PredictResponse{
				Content: scanner.Text(),
			},
		}
		s.sendResponse(response)
	}
	return nil
}

func (s *Server) Predict(ctx context.Context, w io.Writer, text string, providerAndModel string) error {
	parts := strings.Split(providerAndModel, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid provider/model format: must be in format provider/model")
	}
	provider := s.config.Provider
	if provider == nil {
		return fmt.Errorf("provider not set")
	}
	ps := splitter.ProjectContext{Prefix: text}
	_, err := Flow(*provider, ps, ctx, w)
	return err
}
