package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const tempStoreTTL = 60 * time.Second

// HTTPServer handles HTTP API requests for webfetch-clean.
type HTTPServer struct {
	config     Config
	tempStore  *TempStore
	tokenStore *TokenStore
	mux        *http.ServeMux
}

// TempEntry holds an oversized result stored temporarily in memory.
type TempEntry struct {
	Content   string
	Format    string
	URL       string
	CreatedAt time.Time
}

// TempStore is a bounded in-memory store for oversized results.
type TempStore struct {
	mu         sync.RWMutex
	entries    map[string]TempEntry
	maxEntries int
}

// OverLimitResult is the structured JSON returned when output exceeds the token limit.
type OverLimitResult struct {
	ResultID   string `json:"result_id"`
	TokenCount int    `json:"token_count"`
	Limit      int    `json:"limit"`
	Message    string `json:"message"`
}

// AdminTokenRequest is the JSON body for POST /admin/tokens.
type AdminTokenRequest struct {
	File           string `json:"file"`
	ExpiresMinutes int    `json:"expires_minutes"`
}

// AdminTokenResponse is the JSON response from POST /admin/tokens.
type AdminTokenResponse struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

// NewTempStore creates a TempStore with the given capacity.
func NewTempStore(maxEntries int) *TempStore {
	return &TempStore{
		entries:    make(map[string]TempEntry),
		maxEntries: maxEntries,
	}
}

// Store adds an entry. Returns an error if at capacity.
func (ts *TempStore) Store(id string, entry TempEntry) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if len(ts.entries) >= ts.maxEntries {
		return fmt.Errorf("temp store at capacity (%d entries)", ts.maxEntries)
	}
	ts.entries[id] = entry
	return nil
}

// LoadAndDelete retrieves and removes an entry atomically.
func (ts *TempStore) LoadAndDelete(id string) (TempEntry, bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	entry, ok := ts.entries[id]
	if ok {
		delete(ts.entries, id)
	}
	return entry, ok
}

// Cleanup removes entries older than the TTL.
func (ts *TempStore) Cleanup() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	cutoff := time.Now().Add(-tempStoreTTL)
	for id, entry := range ts.entries {
		if entry.CreatedAt.Before(cutoff) {
			delete(ts.entries, id)
		}
	}
}

// Len returns the current number of entries.
func (ts *TempStore) Len() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return len(ts.entries)
}

// newHTTPServer builds an HTTPServer with routes registered.
func newHTTPServer(config Config, tokenStore *TokenStore) *HTTPServer {
	s := &HTTPServer{
		config:     config,
		tempStore:  NewTempStore(100),
		tokenStore: tokenStore,
		mux:        http.NewServeMux(),
	}

	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("POST /mcp", s.requireAuth(s.handleMCP))
	s.mux.HandleFunc("GET /results/{id}", s.requireAuth(s.handleResultDownload))
	s.mux.HandleFunc("POST /admin/tokens", s.requireAuth(s.handleAdminTokens))

	return s
}

// requireAuth wraps a handler with API key authentication.
func (s *HTTPServer) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key == "" {
			http.Error(w, "missing API key", http.StatusUnauthorized)
			return
		}
		if subtle.ConstantTimeCompare([]byte(key), []byte(s.config.APIKey)) != 1 {
			http.Error(w, "invalid API key", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// handleHealth returns a simple health check response.
func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleMCP is the JSON-RPC 2.0 dispatcher for the /mcp endpoint.
func (s *HTTPServer) handleMCP(w http.ResponseWriter, r *http.Request) {
	body := http.MaxBytesReader(w, r.Body, MaxScannerBuffer)
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		s.writeJSONRPCError(w, nil, ErrParse, "request body too large or unreadable")
		return
	}

	var req JSONRPCRequest
	if err := json.Unmarshal(data, &req); err != nil {
		s.writeJSONRPCError(w, nil, ErrParse, "parse error")
		return
	}

	switch req.Method {
	case "initialize":
		s.handleHTTPInitialize(w, req)
	case "notifications/initialized":
		// Accept as no-op (unlike stdio handler which returns MethodNotFound)
		w.WriteHeader(http.StatusNoContent)
	case "tools/list":
		s.handleHTTPToolsList(w, req)
	case "tools/call":
		s.handleHTTPToolsCall(w, req)
	default:
		s.writeJSONRPCError(w, req.ID, ErrMethodNotFound, "method not found")
	}
}

// handleHTTPInitialize returns the MCP initialization response.
func (s *HTTPServer) handleHTTPInitialize(w http.ResponseWriter, req JSONRPCRequest) {
	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		ServerInfo: ServerInfo{
			Name:    "webfetch-clean",
			Version: Version,
		},
		Capabilities: Capabilities{
			Tools: map[string]bool{
				"list": true,
				"call": true,
			},
		},
	}
	s.writeJSONRPCResponse(w, req.ID, result)
}

// handleHTTPToolsList returns the HTTP-specific tool schema.
func (s *HTTPServer) handleHTTPToolsList(w http.ResponseWriter, req JSONRPCRequest) {
	result := ToolsListResult{
		Tools: []Tool{
			{
				Name:        "webfetch_clean",
				Description: "Fetch a URL or read a token-authorized file, clean HTML by removing ads/scripts/styles/navigation, and convert to markdown or cleaned HTML",
				InputSchema: InputSchema{
					Type: "object",
					Properties: map[string]Property{
						"url": {
							Type:        "string",
							Description: "URL to fetch and clean",
						},
						"file_token": {
							Type:        "string",
							Description: "Single-use token for server-side file access (replaces file parameter)",
						},
						"result_id": {
							Type:        "string",
							Description: "ID of a stored over-limit result. Use with override=true to retrieve full content.",
						},
						"override": {
							Type:        "boolean",
							Description: "When true with result_id, returns the full stored content.",
							Default:     false,
						},
						"output_format": {
							Type:        "string",
							Description: "Output format: 'html' or 'markdown' (default: 'markdown')",
							Enum:        []string{"html", "markdown"},
							Default:     "markdown",
						},
						"mode": {
							Type:        "string",
							Description: "Processing mode: 'clean' removes ads/scripts/nav, 'scrape' preserves page structure (default: 'clean')",
							Enum:        []string{"clean", "scrape"},
							Default:     "clean",
						},
						"preserve_main_only": {
							Type:        "boolean",
							Description: "Only preserve content inside <main> or <article> tags (default: false)",
							Default:     false,
						},
						"remove_images": {
							Type:        "boolean",
							Description: "Remove all images from output (default: false)",
							Default:     false,
						},
						"strip_links": {
							Type:        "boolean",
							Description: "Replace links with their text content (default: false)",
							Default:     false,
						},
						"use_browser": {
							Type:        "boolean",
							Description: "Use headless browser for JavaScript-rendered pages. Only valid with 'url'.",
							Default:     false,
						},
						"timeout": {
							Type:        "integer",
							Description: "HTTP request timeout in seconds (default: 30)",
							Default:     30,
						},
						"max_tokens": {
							Type:        "integer",
							Description: "Maximum tokens for output before storing as temp result (default: 100000)",
							Default:     100000,
						},
					},
					Required: []string{},
				},
			},
		},
	}
	s.writeJSONRPCResponse(w, req.ID, result)
}

// handleHTTPToolsCall processes tool calls over HTTP with temp-store over-limit handling.
func (s *HTTPServer) handleHTTPToolsCall(w http.ResponseWriter, req JSONRPCRequest) {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.writeJSONRPCError(w, req.ID, ErrInvalidParams, "invalid params")
		return
	}

	if params.Name != "webfetch_clean" {
		s.writeJSONRPCError(w, req.ID, ErrInvalidParams, "unknown tool")
		return
	}

	// Reject raw file parameter in HTTP mode
	if _, hasFile := params.Arguments["file"]; hasFile {
		s.writeJSONRPCError(w, req.ID, ErrInvalidParams,
			"'file' parameter is not available in HTTP mode. Use 'file_token' instead.")
		return
	}

	// Handle result_id + override path (retrieve stored over-limit content)
	if resultID, ok := params.Arguments["result_id"].(string); ok && resultID != "" {
		override, _ := params.Arguments["override"].(bool)
		if !override {
			s.writeJSONRPCError(w, req.ID, ErrInvalidParams,
				"result_id requires override=true to retrieve stored content")
			return
		}
		entry, found := s.tempStore.LoadAndDelete(resultID)
		if !found {
			s.writeJSONRPCError(w, req.ID, ErrInvalidParams,
				"result not found or expired")
			return
		}
		result := CleanResult{
			Content: entry.Content,
			URL:     entry.URL,
			Format:  entry.Format,
		}
		s.writeToolCallResult(w, req.ID, result)
		return
	}

	url, _ := params.Arguments["url"].(string)
	fileToken, _ := params.Arguments["file_token"].(string)

	if url != "" && fileToken != "" {
		s.writeJSONRPCError(w, req.ID, ErrInvalidParams,
			"cannot specify both 'url' and 'file_token' parameters")
		return
	}
	if url == "" && fileToken == "" {
		s.writeJSONRPCError(w, req.ID, ErrInvalidParams,
			"must provide 'url' or 'file_token' parameter")
		return
	}

	config := Config{
		URL:       url,
		Format:    "markdown",
		Mode:      "clean",
		Timeout:   30,
		MaxTokens: 100000,
	}

	// Resolve file_token to file path
	if fileToken != "" {
		filePath, err := s.tokenStore.RedeemToken(fileToken)
		if err != nil {
			s.writeJSONRPCError(w, req.ID, ErrInvalidParams,
				fmt.Sprintf("file token error: %v", err))
			return
		}
		config.File = filePath
	}

	if format, ok := params.Arguments["output_format"].(string); ok {
		config.Format = format
	}
	if mode, ok := params.Arguments["mode"].(string); ok {
		config.Mode = mode
	}
	if preserveMain, ok := params.Arguments["preserve_main_only"].(bool); ok {
		config.PreserveMain = preserveMain
	}
	if removeImages, ok := params.Arguments["remove_images"].(bool); ok {
		config.RemoveImages = removeImages
	}
	if stripLinks, ok := params.Arguments["strip_links"].(bool); ok {
		config.StripLinks = stripLinks
	}
	if useBrowser, ok := params.Arguments["use_browser"].(bool); ok {
		config.UseBrowser = useBrowser
	}
	if config.UseBrowser && config.File != "" {
		s.writeJSONRPCError(w, req.ID, ErrInvalidParams,
			"cannot use 'use_browser' with file token")
		return
	}
	if timeout, ok := params.Arguments["timeout"].(float64); ok {
		config.Timeout = int(timeout)
	}
	if maxTokens, ok := params.Arguments["max_tokens"].(float64); ok {
		config.MaxTokens = int(maxTokens)
	}

	result := processInput(config)

	if result.Error != "" {
		s.writeToolCallResult(w, req.ID, result)
		return
	}

	// Handle over-limit: store in temp store
	if result.OverLimit {
		id := generateUUID()
		entry := TempEntry{
			Content:   result.RawContent,
			Format:    result.Format,
			URL:       result.URL,
			CreatedAt: time.Now(),
		}
		storeErr := s.tempStore.Store(id, entry)
		if storeErr != nil {
			// Capacity full — truncate and return what we can
			result.Content = result.RawContent[:config.MaxTokens*3]
			result.RawContent = ""
			s.writeToolCallResult(w, req.ID, result)
			return
		}

		baseURL := s.config.BaseURL
		ttlSeconds := int(tempStoreTTL.Seconds())

		overLimit := OverLimitResult{
			ResultID:   id,
			TokenCount: result.TokenCount,
			Limit:      config.MaxTokens,
			Message: fmt.Sprintf(
				"Fetched %s — %d tokens (limit: %d).\nResult stored for %d seconds (ID: %s).\n\nOptions:\n1. Override limit: call webfetch_clean with {\"result_id\": \"%s\", \"override\": true}\n2. Download: GET %s/results/%s\n3. Do nothing — result expires in %d seconds.",
				result.URL, result.TokenCount, config.MaxTokens,
				ttlSeconds, id, id, baseURL, id, ttlSeconds),
		}

		overLimitJSON, err := json.Marshal(overLimit)
		if err != nil {
			s.writeJSONRPCError(w, req.ID, ErrInternal, "failed to marshal over-limit result")
			return
		}

		response := ToolCallResult{
			Content: []ContentItem{
				{
					Type: "text",
					Text: string(overLimitJSON),
				},
			},
		}
		s.writeJSONRPCResponse(w, req.ID, response)
		return
	}

	s.writeToolCallResult(w, req.ID, result)
}

// handleResultDownload serves a stored over-limit result and deletes it.
func (s *HTTPServer) handleResultDownload(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	entry, found := s.tempStore.LoadAndDelete(id)
	if !found {
		http.Error(w, "result not found or expired", http.StatusNotFound)
		return
	}

	contentType := "text/markdown; charset=utf-8"
	ext := ".md"
	if entry.Format == "html" {
		contentType = "text/html; charset=utf-8"
		ext = ".html"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"result-%s%s\"", id[:8], ext))
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, entry.Content)
}

// handleAdminTokens creates a file access token.
func (s *HTTPServer) handleAdminTokens(w http.ResponseWriter, r *http.Request) {
	body := http.MaxBytesReader(w, r.Body, MaxScannerBuffer)
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		return
	}

	var req AdminTokenRequest
	if err := json.Unmarshal(data, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.File == "" {
		http.Error(w, "file path is required", http.StatusBadRequest)
		return
	}

	if req.ExpiresMinutes <= 0 {
		req.ExpiresMinutes = 60
	}
	if req.ExpiresMinutes > 1440 {
		http.Error(w, "expires_minutes must be between 1 and 1440", http.StatusBadRequest)
		return
	}

	ttl := time.Duration(req.ExpiresMinutes) * time.Minute
	token, err := s.tokenStore.CreateFileToken(req.File, ttl)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create token: %v", err), http.StatusBadRequest)
		return
	}

	resp := AdminTokenResponse{
		Token:   token,
		Expires: time.Now().Add(ttl),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// writeJSONRPCResponse writes a successful JSON-RPC response.
func (s *HTTPServer) writeJSONRPCResponse(w http.ResponseWriter, id any, result any) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// writeJSONRPCError writes a JSON-RPC error response.
func (s *HTTPServer) writeJSONRPCError(w http.ResponseWriter, id any, code int, message string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// writeToolCallResult marshals a CleanResult into a JSON-RPC tools/call response.
func (s *HTTPServer) writeToolCallResult(w http.ResponseWriter, id any, result CleanResult) {
	jsonResult, err := json.Marshal(result)
	if err != nil {
		s.writeJSONRPCError(w, id, ErrInternal, "failed to marshal result")
		return
	}

	response := ToolCallResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: string(jsonResult),
			},
		},
	}
	s.writeJSONRPCResponse(w, id, response)
}

// runHTTPServer starts the HTTP server with graceful shutdown.
func runHTTPServer(config Config) {
	// Resolve API key: flag > env var
	if config.APIKey == "" {
		config.APIKey = os.Getenv("WEBFETCH_API_KEY")
	}
	if config.APIKey == "" {
		fmt.Fprintln(os.Stderr, "Error: API key required for HTTP mode (use --api-key or WEBFETCH_API_KEY env)")
		os.Exit(ExitInvalidArgs)
	}

	// Normalize base URL
	config.BaseURL = strings.TrimRight(config.BaseURL, "/")

	// Default bind to loopback
	addr := config.HTTPAddr
	if strings.HasPrefix(addr, ":") {
		addr = "127.0.0.1" + addr
	}

	tokenStore, err := NewTokenStore(config.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening token database: %v\n", err)
		os.Exit(ExitProcessError)
	}
	defer tokenStore.Close()

	srv := newHTTPServer(config, tokenStore)

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      srv.mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Temp store cleanup goroutine
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				srv.tempStore.Cleanup()
			}
		}
	}()

	// Token store cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := tokenStore.Cleanup(); err != nil {
					log.Printf("token cleanup error: %v", err)
				}
			}
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("shutting down HTTP server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		cancel() // stop cleanup goroutines

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown error: %v", err)
		}

		// Second signal = immediate exit
		<-sigChan
		os.Exit(1)
	}()

	log.Printf("HTTP server listening on %s", addr)
	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
		os.Exit(ExitProcessError)
	}
}
