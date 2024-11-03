package lsp

// Position in a text document expressed as zero-based line and character offset
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range in a text document expressed as (start, end) positions
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// TextDocumentItem represents an open text document
type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

// InitializeParams represents parameters for initialize request
type InitializeParams struct {
	ProcessID int    `json:"processId"`
	RootURI   string `json:"rootUri"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult represents result of initialize request
type InitializeResult struct {
	Info         ServerInfo         `json:"serverInfo"`
	Capabilities ServerCapabilities `json:"capabilities"`
}

// TextDocumentSyncOptions represents synchronization options
type TextDocumentSyncOptions struct {
	OpenClose bool        `json:"openClose"`
	Change    int         `json:"change"`
	Save      SaveOptions `json:"save"`
}

// SaveOptions represents options for document save
type SaveOptions struct {
	IncludeText bool `json:"includeText"`
}

// ServerCapabilities represents server capabilities
type ServerCapabilities struct {
	TextDocumentSync   TextDocumentSyncOptions `json:"textDocumentSync"`
	CompletionProvider bool                    `json:"completionProvider"`
}

// DidOpenTextDocumentParams params for textDocument/didOpen
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// TextDocumentContentChangeEvent represents a change to a text document
type TextDocumentContentChangeEvent struct {
	Range Range  `json:"range"`
	Text  string `json:"text"`
}

// DidChangeTextDocumentParams params for textDocument/didChange
type DidChangeTextDocumentParams struct {
	TextDocument   TextDocumentItem                 `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// CompletionItem represents a completion item
type CompletionItem struct {
	Label  string `json:"label"`
	Kind   int    `json:"kind"`
	Detail string `json:"detail"`
}

// CompletionList represents a list of completion items
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// PredictParams represents parameters for predict request
type PredictParams struct {
	Text string `json:"text"`
}

// PredictResponse represents a single prediction response
type PredictResponse struct {
	ID      interface{} `json:"id"`
	Content string      `json:"content"`
}
