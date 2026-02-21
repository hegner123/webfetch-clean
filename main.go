package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

// Version is the application version
const Version = "1.0.0"

// MaxScannerBuffer is the maximum buffer size for reading MCP requests (10MB)
const MaxScannerBuffer = 10 * 1024 * 1024

// Exit codes for CLI mode
const (
	ExitSuccess        = 0
	ExitInvalidArgs    = 1
	ExitFetchError     = 2
	ExitProcessError   = 3
	ExitFileWriteError = 4
)

// JSON-RPC 2.0 error codes
const (
	ErrParse          = -32700
	ErrInvalidRequest = -32600
	ErrMethodNotFound = -32601
	ErrInvalidParams  = -32602
	ErrInternal       = -32603
)

// Config holds the application configuration
type Config struct {
	URL             string
	File            string
	Format          string
	Mode            string
	PreserveMain    bool
	RemoveImages    bool
	StripLinks      bool
	UseBrowser      bool
	Timeout         int
	OutputFile      string
	OutputDirectory string
	MaxTokens       int
	CLIMode         bool
	ShowVersion     bool
	Verbose         bool
	HTTPAddr        string
	APIKey          string
	BaseURL         string
	DBPath          string
}

// CleanResult represents the result of cleaning a URL
type CleanResult struct {
	Content    string `json:"content"`
	URL        string `json:"url"`
	Title      string `json:"title,omitempty"`
	Format     string `json:"format"`
	Error      string `json:"error,omitempty"`
	TokenCount int    `json:"token_count,omitempty"`
	OverLimit  bool   `json:"-"`
	RawContent string `json:"-"`
}

// MCP JSON-RPC types
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type InitializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	ServerInfo      ServerInfo   `json:"serverInfo"`
	Capabilities    Capabilities `json:"capabilities"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Capabilities struct {
	Tools map[string]bool `json:"tools"`
}

type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
	Default     any      `json:"default,omitempty"`
}

type ToolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type ToolCallResult struct {
	Content []ContentItem `json:"content"`
}

type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func main() {
	config := parseFlags()

	if config.ShowVersion {
		fmt.Printf("webfetch-clean version %s\n", Version)
		return
	}

	if config.HTTPAddr != "" {
		runHTTPServer(config)
		return
	}

	if config.CLIMode {
		runCLI(config)
		return
	}

	runMCPServer()
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `webfetch-clean - Fetch and clean web pages

USAGE:
  webfetch-clean [options]

DESCRIPTION:
  In CLI mode, fetches a URL, removes ads/scripts/navigation, and outputs
  clean HTML or Markdown. By default, runs as an MCP server.

MCP MODE (default):
  Runs a JSON-RPC 2.0 MCP server on stdin/stdout for Claude Code integration.

CLI MODE:
  webfetch-clean --cli --url https://example.com
  webfetch-clean --cli --file local.html
  webfetch-clean --cli --url https://example.com --format html
  webfetch-clean --cli --url https://example.com --output result.md

HTTP MODE:
  webfetch-clean --http :8080 --api-key SECRET --base-url http://localhost:8080
  Endpoints: POST /mcp, GET /results/{id}, POST /admin/tokens, GET /health

OPTIONS:
`)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
EXAMPLES:
  # Fetch a page and output markdown to stdout
  webfetch-clean --cli --url https://example.com

  # Save cleaned HTML to a file
  webfetch-clean --cli --url https://example.com --format html --output page.html

  # Keep only main content, remove images
  webfetch-clean --cli --url https://example.com --preserve-main --remove-images

  # Use headless browser for JavaScript-rendered pages
  webfetch-clean --cli --url https://spa-example.com --browser

  # Run as HTTP server with API key auth
  webfetch-clean --http :8080 --api-key my-secret --base-url http://localhost:8080

  # Run as MCP server (default)
  webfetch-clean
`)
}

func parseFlags() Config {
	config := Config{}

	flag.Usage = printUsage

	flag.BoolVar(&config.ShowVersion, "version", false, "Show version and exit")
	flag.BoolVar(&config.CLIMode, "cli", false, "Run in CLI mode (default: MCP server mode)")
	flag.StringVar(&config.URL, "url", "", "URL to fetch (alternative to --file)")
	flag.StringVar(&config.File, "file", "", "Local HTML file to process (alternative to --url)")
	flag.StringVar(&config.Format, "format", "markdown", "Output format: html or markdown (CLI mode only)")
	flag.BoolVar(&config.PreserveMain, "preserve-main", false, "Only preserve <main>/<article> content (CLI mode only)")
	flag.BoolVar(&config.RemoveImages, "remove-images", false, "Remove all images (CLI mode only)")
	flag.StringVar(&config.Mode, "mode", "clean", "Processing mode: clean or scrape (CLI mode only)")
	flag.BoolVar(&config.StripLinks, "strip-links", false, "Replace links with their text content (CLI mode only)")
	flag.BoolVar(&config.UseBrowser, "browser", false, "Use headless browser for JavaScript-rendered pages (CLI mode only)")
	flag.IntVar(&config.Timeout, "timeout", 30, "HTTP timeout in seconds (CLI mode only)")
	flag.StringVar(&config.OutputFile, "output", "", "Write output to file (default: stdout, CLI mode only)")
	flag.BoolVar(&config.Verbose, "verbose", false, "Print verbose progress messages to stderr (CLI mode only)")

	// HTTP server mode flags
	flag.StringVar(&config.HTTPAddr, "http", "", "Start HTTP server on address (e.g., :8080)")
	flag.StringVar(&config.APIKey, "api-key", "", "API key for HTTP mode (or WEBFETCH_API_KEY env)")
	flag.StringVar(&config.BaseURL, "base-url", "", "Public base URL for download links (e.g., https://fetch.example.com)")
	flag.StringVar(&config.DBPath, "db", "webfetch.db", "SQLite database path for file tokens (HTTP mode)")

	flag.Parse()

	return config
}

func runCLI(config Config) {
	if config.URL == "" && config.File == "" {
		fmt.Fprintln(os.Stderr, "Error: --url or --file is required")
		flag.Usage()
		os.Exit(ExitInvalidArgs)
	}

	if config.URL != "" && config.File != "" {
		fmt.Fprintln(os.Stderr, "Error: cannot specify both --url and --file")
		os.Exit(ExitInvalidArgs)
	}

	if config.UseBrowser && config.File != "" {
		fmt.Fprintln(os.Stderr, "Error: --browser cannot be used with --file")
		os.Exit(ExitInvalidArgs)
	}

	if config.Format != "html" && config.Format != "markdown" {
		fmt.Fprintln(os.Stderr, "Error: --format must be 'html' or 'markdown'")
		os.Exit(ExitInvalidArgs)
	}

	if config.Mode != "clean" && config.Mode != "scrape" {
		fmt.Fprintln(os.Stderr, "Error: --mode must be 'clean' or 'scrape'")
		os.Exit(ExitInvalidArgs)
	}

	if config.Verbose {
		if config.URL != "" {
			fmt.Fprintf(os.Stderr, "[verbose] Fetching URL: %s\n", config.URL)
			fmt.Fprintf(os.Stderr, "[verbose] Timeout: %d seconds\n", config.Timeout)
			if config.UseBrowser {
				fmt.Fprintln(os.Stderr, "[verbose] Using headless browser (Chromium) for JS-rendered content")
			}
		} else {
			fmt.Fprintf(os.Stderr, "[verbose] Reading file: %s\n", config.File)
		}
		fmt.Fprintf(os.Stderr, "[verbose] Output format: %s\n", config.Format)
		fmt.Fprintf(os.Stderr, "[verbose] Processing mode: %s\n", config.Mode)
		if config.PreserveMain {
			fmt.Fprintln(os.Stderr, "[verbose] Preserving main/article content only")
		}
		if config.RemoveImages {
			fmt.Fprintln(os.Stderr, "[verbose] Removing images")
		}
		if config.StripLinks {
			fmt.Fprintln(os.Stderr, "[verbose] Stripping links")
		}
	}

	result := processInput(config)

	if result.Error != "" {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.Error)
		os.Exit(ExitProcessError)
	}

	if result.OverLimit {
		filename := GenerateOutputFilename(config.URL, config.File, config.Format)
		if config.OutputDirectory != "" {
			filename = filepath.Join(config.OutputDirectory, filename)
		}
		filename, err := GenerateUniqueFilename(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating unique filename: %v\n", err)
			os.Exit(ExitFileWriteError)
		}
		err = SafeWriteFile(filename, []byte(result.RawContent), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
			os.Exit(ExitFileWriteError)
		}
		result.Content = fmt.Sprintf("Output exceeded limit (%d tokens, limit: %d tokens). Content written to file: %s",
			result.TokenCount, config.MaxTokens, filename)
		result.RawContent = ""
	}

	output := result.Content

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "[verbose] Content fetched and cleaned successfully\n")
		fmt.Fprintf(os.Stderr, "[verbose] Output size: %d bytes\n", len(output))
	}

	if config.OutputFile != "" {
		if config.Verbose {
			fmt.Fprintf(os.Stderr, "[verbose] Writing to file: %s\n", config.OutputFile)
		}
		err := SafeWriteFile(config.OutputFile, []byte(output), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
			os.Exit(ExitFileWriteError)
		}
		if config.Verbose {
			fmt.Fprintf(os.Stderr, "[verbose] File written successfully\n")
		}
	} else {
		fmt.Println(output)
	}
}

func runMCPServer() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Create a channel to receive lines from stdin
	lines := make(chan string)
	scanErr := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 0, MaxScannerBuffer), MaxScannerBuffer)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		scanErr <- scanner.Err()
		close(lines)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case line, ok := <-lines:
			if !ok {
				// stdin closed
				return
			}
			if line == "" {
				continue
			}

			var req JSONRPCRequest
			if err := json.Unmarshal([]byte(line), &req); err != nil {
				sendError(nil, ErrParse, "Parse error")
				continue
			}

			handleRequest(req)
		}
	}
}

func handleRequest(req JSONRPCRequest) {
	// JSON-RPC 2.0: notifications (no id) must not receive a response
	isNotification := req.ID == nil

	switch req.Method {
	case "initialize":
		handleInitialize(req)
	case "notifications/initialized":
		// MCP lifecycle notification, no response required
		return
	case "tools/list":
		handleToolsList(req)
	case "tools/call":
		handleToolsCall(req)
	default:
		if isNotification {
			return
		}
		sendError(req.ID, ErrMethodNotFound, "Method not found")
	}
}

func handleInitialize(req JSONRPCRequest) {
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
	sendResponse(req.ID, result)
}

func handleToolsList(req JSONRPCRequest) {
	result := ToolsListResult{
		Tools: []Tool{
			{
				Name:        "webfetch_clean",
				Description: "Fetch a URL or read a local file, clean HTML by removing ads/scripts/styles/navigation, and convert to markdown or cleaned HTML",
				InputSchema: InputSchema{
					Type: "object",
					Properties: map[string]Property{
						"url": {
							Type:        "string",
							Description: "URL to fetch and clean (provide url or file, not both)",
						},
						"file": {
							Type:        "string",
							Description: "Local file path to read and clean (provide url or file, not both)",
						},
						"output_format": {
							Type:        "string",
							Description: "Output format: 'html' or 'markdown' (default: 'markdown')",
							Enum:        []string{"html", "markdown"},
							Default:     "markdown",
						},
						"mode": {
							Type:        "string",
							Description: "Processing mode: 'clean' removes ads/scripts/nav for AI workflows, 'scrape' preserves page structure for HTML ingestion (default: 'clean')",
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
							Description: "Use headless browser (Chromium) for JavaScript-rendered pages. Only valid with 'url'. Slower but handles SPAs.",
							Default:     false,
						},
						"timeout": {
							Type:        "integer",
							Description: "HTTP request timeout in seconds (default: 30)",
							Default:     30,
						},
						"output_directory": {
							Type:        "string",
							Description: "Directory to write output files when token limit is exceeded (default: current directory)",
							Default:     "",
						},
						"max_tokens": {
							Type:        "integer",
							Description: "Maximum tokens for output before writing to file (default: 100000, where 3 bytes = 1 token)",
							Default:     100000,
						},
					},
					Required: []string{},
				},
			},
		},
	}
	sendResponse(req.ID, result)
}

func handleToolsCall(req JSONRPCRequest) {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		sendError(req.ID, ErrInvalidParams, "Invalid params")
		return
	}

	if params.Name != "webfetch_clean" {
		sendError(req.ID, ErrInvalidParams, "Unknown tool")
		return
	}

	url, _ := params.Arguments["url"].(string)
	file, _ := params.Arguments["file"].(string)

	if url != "" && file != "" {
		sendError(req.ID, ErrInvalidParams, "Cannot specify both 'url' and 'file' parameters")
		return
	}
	if url == "" && file == "" {
		sendError(req.ID, ErrInvalidParams, "Must provide either 'url' or 'file' parameter")
		return
	}

	config := Config{
		URL:       url,
		File:      file,
		Format:    "markdown",
		Mode:      "clean",
		Timeout:   30,
		MaxTokens: 100000,
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
		sendError(req.ID, ErrInvalidParams, "Cannot use 'use_browser' with 'file' parameter")
		return
	}

	if timeout, ok := params.Arguments["timeout"].(float64); ok {
		config.Timeout = int(timeout)
	}

	if outputDir, ok := params.Arguments["output_directory"].(string); ok {
		config.OutputDirectory = outputDir
	}

	if maxTokens, ok := params.Arguments["max_tokens"].(float64); ok {
		config.MaxTokens = int(maxTokens)
	}

	result := processInput(config)

	if result.OverLimit {
		filename := GenerateOutputFilename(config.URL, config.File, config.Format)
		if config.OutputDirectory != "" {
			filename = filepath.Join(config.OutputDirectory, filename)
		}
		filename, err := GenerateUniqueFilename(filename)
		if err != nil {
			sendError(req.ID, ErrInternal, fmt.Sprintf("failed to generate unique filename: %v", err))
			return
		}
		err = SafeWriteFile(filename, []byte(result.RawContent), 0644)
		if err != nil {
			sendError(req.ID, ErrInternal, fmt.Sprintf("failed to write output file: %v", err))
			return
		}
		result.Content = fmt.Sprintf("Output exceeded limit (%d tokens, limit: %d tokens). Content written to file: %s",
			result.TokenCount, config.MaxTokens, filename)
		result.RawContent = ""
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		sendError(req.ID, ErrInternal, "Failed to marshal result")
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

	sendResponse(req.ID, response)
}

func processInput(config Config) CleanResult {
	result := CleanResult{
		URL:    config.URL,
		Format: config.Format,
	}

	var html string
	var err error

	// Step 1: Get HTML from URL or file
	if config.File != "" {
		html, err = ReadFile(config.File)
	} else if config.UseBrowser {
		var finalURL string
		html, finalURL, err = FetchBrowser(context.Background(), config.URL, config.Timeout)
		if finalURL != "" {
			result.URL = finalURL
		}
	} else {
		var finalURL string
		html, finalURL, err = FetchURL(context.Background(), config.URL, config.Timeout)
		if finalURL != "" {
			result.URL = finalURL
		}
	}
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// Step 2: Clean the HTML
	cleanedHTML, err := CleanHTML(html, config.PreserveMain, config.RemoveImages, config.StripLinks, config.Mode)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// Step 3: Convert to requested format
	output, err := ConvertToFormat(cleanedHTML, config.Format)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// Step 4: Check token limit — signal over-limit to caller, no I/O here
	// Token calculation: 3 bytes = 1 token
	tokenCount := len(output) / 3
	if config.MaxTokens > 0 && tokenCount > config.MaxTokens {
		result.OverLimit = true
		result.TokenCount = tokenCount
		result.RawContent = output
		return result
	}

	result.Content = output
	return result
}

func sendResponse(id any, result any) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal response: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func sendError(id any, code int, message string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal error response: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

// SafeWriteFile writes data to a file, refusing to write through symlinks
// to prevent symlink attacks. If the path is a symlink, it returns an error.
func SafeWriteFile(path string, data []byte, perm os.FileMode) error {
	// Check if path exists and is a symlink
	info, err := os.Lstat(path)
	if err == nil {
		// Path exists - check if it's a symlink
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing to write to symlink: %s", path)
		}
	} else if !os.IsNotExist(err) {
		// Lstat failed for a reason other than "not exist"
		return fmt.Errorf("failed to check path: %w", err)
	}
	// If path doesn't exist or is a regular file, proceed with write
	return os.WriteFile(path, data, perm)
}
