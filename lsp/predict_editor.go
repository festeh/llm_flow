package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/festeh/llm_flow/lsp/splitter"
)

func (s *Server) HandlePredictEditor(header Header, ctx context.Context) error {
	var params PredictEditorParams
	if err := json.Unmarshal(header.Params, &params); err != nil {
		return fmt.Errorf("invalid predict params: %v", err)
	}
	log.Info("got predict_request", "id", header.ID, "line", params.Line, "pos", params.Pos)
	// Create cancellable context
	predCtx, cancel := context.WithCancel(ctx)
	s.predictionsMu.Lock()
	s.activePredictions[header.ID] = cancel
	s.predictionsMu.Unlock()

	_, pw := io.Pipe()
	go func() {
		defer pw.Close()
		content, err := s.PredictEditor(predCtx, pw, params)
		// Clean up prediction tracking
		if err != nil {
			log.Error("Prediction", "error", err, "id", header.ID)
			s.predictionsMu.Lock()
			s.sendCancel(header.ID)
			s.predictionsMu.Unlock()
			return
		}
		log.Info("Done", "id", header.ID)
		// Send completion notification after prediction is done
		response := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      header.ID,
			"result": PredictResponse{
				ID:      header.ID,
				Content: content,
			},
		}
		s.predictionsMu.Lock()
		delete(s.activePredictions, header.ID)
		s.sendResponse(response)
		s.predictionsMu.Unlock()
	}()

	// scanner := bufio.NewScanner(pr)
	// for scanner.Scan() {
	// _ := map[string]interface{}{
	// 	"jsonrpc": "2.0",
	// 	"method":  "predict/response",
	// 	"params": PredictResponse{
	// 		Content: scanner.Text(),
	// 	},
	// }
	// s.sendResponse(response)
	// }
	return nil
}

func (s *Server) PredictEditor(ctx context.Context, w io.Writer, params PredictEditorParams) (string, error) {
	if s.config.Provider == nil {
		return "", fmt.Errorf("Provider not set")
	}
	// Get document content
	doc, exists := s.documents[params.URI]
	if !exists {
		return "", fmt.Errorf("document not found: %s", params.URI)
	}

	// Split document into lines
	lines := strings.Split(doc, "\n")
	if params.Line >= len(lines) {
		return "", fmt.Errorf("line number out of range: %d", params.Line)
	}

	currentLine := lines[params.Line]
	prefix := strings.Join(lines[:params.Line], "\n")
	if params.Line > 0 {
		prefix += "\n"
	}

	pos := params.Pos
	suffix := ""
	if pos >= len(currentLine) {
		prefix += currentLine
	} else {
		prefix += currentLine[:pos]
		suffix = currentLine[pos:]
	}

	if params.Line < len(lines)-1 {
		suffix += "\n" + strings.Join(lines[params.Line+1:], "\n")
	}
	filePath := strings.TrimPrefix(params.URI, "file://")
  prefixSuffix := splitter.ProjectContext{Repo: s.config.Repo, Prefix: prefix, Suffix: suffix, File: filePath}
	return Flow(*s.config.Provider, prefixSuffix, ctx, w)
}

func (s *Server) HandleCancelPredictEditor(header Header) {
	s.predictionsMu.Lock()
	params := CancelParams{}
	err := json.Unmarshal(header.Params, &params)
	if err != nil {
		log.Info("Err in cancel", err)
	}
	id := params.ID
	log.Info("Cancel", "id", id)
	cancel, ok := s.activePredictions[id]
	if ok {
		cancel()
		delete(s.activePredictions, id)
		log.Info("Cancelled prediction %v", id)
	}
	s.predictionsMu.Unlock()
}
