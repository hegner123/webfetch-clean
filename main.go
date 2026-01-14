package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// Config holds the application configuration
type Config struct {
	URL             string
	Format          string
	PreserveMain    bool
	RemoveImages    bool
	Timeout         int
	OutputFile      string
	CLIMode         bool
}

// CleanResult represents the result of cleaning a URL
type CleanResult struct {
	Content string            `json:"content"`
	URL     string            `json:"url"`
	Title   string            `json:"title,omitempty"`
	Format  string            `json:"format"`
	Error   string            `json:"error,omitempty"`
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

	if config.CLIMode {
		runCLI(config)
		return
	}

	runMCPServer()
}

func parseFlags() Config {
	config := Config{}

	flag.BoolVar(&config.CLIMode, "cli", false, "Run in CLI mode (default: MCP server mode)")
	flag.StringVar(&config.URL, "url", "", "URL to fetch (required in CLI mode)")
	flag.StringVar(&config.Format, "format", "markdown", "Output format: html or markdown (CLI mode only)")
	flag.BoolVar(&config.PreserveMain, "preserve-main", false, "Only preserve <main>/<article> content (CLI mode only)")
	flag.BoolVar(&config.RemoveImages, "remove-images", false, "Remove all images (CLI mode only)")
	flag.IntVar(&config.Timeout, "timeout", 30, "HTTP timeout in seconds (CLI mode only)")
	flag.StringVar(&config.OutputFile, "output", "", "Write output to file (default: stdout, CLI mode only)")

	flag.Parse()

	return config
}

func runCLI(config Config) {
	if config.URL == "" {
		fmt.Fprintln(os.Stderr, "Error: --url is required")
		flag.Usage()
		os.Exit(1)
	}

	if config.Format != "html" && config.Format != "markdown" {
		fmt.Fprintln(os.Stderr, "Error: --format must be 'html' or 'markdown'")
		os.Exit(1)
	}

	result := processURL(config)

	if result.Error != "" {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.Error)
		os.Exit(1)
	}

	output := result.Content

	if config.OutputFile != "" {
		err := os.WriteFile(config.OutputFile, []byte(output), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println(output)
	}
}

func runMCPServer() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			sendError(nil, -32700, "Parse error")
			continue
		}

		handleRequest(req)
	}
}

func handleRequest(req JSONRPCRequest) {
	switch req.Method {
	case "initialize":
		handleInitialize(req)
	case "tools/list":
		handleToolsList(req)
	case "tools/call":
		handleToolsCall(req)
	default:
		sendError(req.ID, -32601, "Method not found")
	}
}

func handleInitialize(req JSONRPCRequest) {
	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		ServerInfo: ServerInfo{
			Name:    "webfetch-clean",
			Version: "1.0.0",
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
				Description: "Fetch a URL, clean HTML by removing ads/scripts/styles/navigation, and convert to markdown or cleaned HTML",
				InputSchema: InputSchema{
					Type: "object",
					Properties: map[string]Property{
						"url": {
							Type:        "string",
							Description: "URL to fetch and clean (required)",
						},
						"output_format": {
							Type:        "string",
							Description: "Output format: 'html' or 'markdown' (default: 'markdown')",
							Enum:        []string{"html", "markdown"},
							Default:     "markdown",
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
						"timeout": {
							Type:        "integer",
							Description: "HTTP request timeout in seconds (default: 30)",
							Default:     30,
						},
					},
					Required: []string{"url"},
				},
			},
		},
	}
	sendResponse(req.ID, result)
}

func handleToolsCall(req JSONRPCRequest) {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		sendError(req.ID, -32602, "Invalid params")
		return
	}

	if params.Name != "webfetch_clean" {
		sendError(req.ID, -32602, "Unknown tool")
		return
	}

	url, ok := params.Arguments["url"].(string)
	if !ok || url == "" {
		sendError(req.ID, -32602, "Missing or invalid 'url' parameter")
		return
	}

	config := Config{
		URL:     url,
		Format:  "markdown",
		Timeout: 30,
	}

	if format, ok := params.Arguments["output_format"].(string); ok {
		config.Format = format
	}

	if preserveMain, ok := params.Arguments["preserve_main_only"].(bool); ok {
		config.PreserveMain = preserveMain
	}

	if removeImages, ok := params.Arguments["remove_images"].(bool); ok {
		config.RemoveImages = removeImages
	}

	if timeout, ok := params.Arguments["timeout"].(float64); ok {
		config.Timeout = int(timeout)
	}

	result := processURL(config)

	jsonResult, err := json.Marshal(result)
	if err != nil {
		sendError(req.ID, -32603, "Failed to marshal result")
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

func processURL(config Config) CleanResult {
	result := CleanResult{
		URL:    config.URL,
		Format: config.Format,
	}

	// Step 1: Fetch the URL
	html, err := FetchURL(config.URL, config.Timeout)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// Step 2: Clean the HTML
	cleanedHTML, err := CleanHTML(html, config.PreserveMain, config.RemoveImages)
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
