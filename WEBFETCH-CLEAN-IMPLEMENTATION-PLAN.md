# webfetch-clean Implementation Plan

## Overview
MCP tool that fetches URLs, removes HTML clutter (ads, scripts, nav, etc.), and outputs cleaned HTML or Markdown. Built in Go with dual-mode support (CLI + MCP server).

## Project Details
- **Language:** Go 1.23
- **Pattern:** Dual-mode (CLI + MCP server with `--mcp` flag)
- **Transport:** stdio (stdin/stdout JSON-RPC 2.0)
- **Reference:** Based on `/Users/home/Documents/Code/Go_dev/checkfor/main.go` pattern

## Architecture

### File Structure
```
webfetch-clean/
├── main.go                              # Entry point, MCP protocol, CLI routing
├── fetcher.go                           # HTTP client with timeout/error handling
├── cleaner.go                           # Multi-pass HTML cleaning pipeline
├── converter.go                         # HTML-to-Markdown conversion
├── go.mod                               # Dependencies
├── .mcp.json                            # MCP server configuration
├── WEBFETCH-CLEAN-IMPLEMENTATION-PLAN.md # This file
├── START.md                             # Onboarding and step tracking
└── README.md                            # User documentation
```

### Dependencies
```go
require (
    github.com/PuerkitoBio/goquery v1.9.0                   // HTML parsing
    github.com/JohannesKaufmann/html-to-markdown v1.6.0    // Markdown conversion
)
```

### MCP Tool Schema
```json
{
  "name": "webfetch_clean",
  "description": "Fetch URL, clean HTML (remove ads/scripts/nav), convert to markdown or HTML",
  "inputSchema": {
    "properties": {
      "url": {"type": "string", "required": true},
      "output_format": {"type": "string", "enum": ["html", "markdown"], "default": "markdown"},
      "preserve_main_only": {"type": "boolean", "default": false},
      "remove_images": {"type": "boolean", "default": false},
      "timeout": {"type": "integer", "default": 30}
    }
  }
}
```

## HTML Cleaning Pipeline

### Multi-Pass Strategy (using goquery)
1. **Remove noise:** `<head>`, `<script>`, `<style>`, `<nav>`
2. **Remove ads:** Elements with class/id containing: `ad`, `advertisement`, `banner`
3. **Remove tracking:** `<iframe>` elements
4. **Remove clutter:** `<footer>`, `<aside>`, elements with `sidebar`, `menu`, `popup`, `modal`, `cookie`
5. **Strip attributes:** Remove all inline attributes except `href`, `src`, `alt`, `title`
6. **Preserve semantic:** `<main>`, `<article>`, `<p>`, `<h1-h6>`, `<ul>`, `<ol>`, `<code>`, `<pre>`, `<table>`, `<a>`, `<img>`

### Implementation Example
```go
// Pass 1: Remove entire sections
doc.Find("head, script, style, nav").Remove()

// Pass 2: Remove ads
doc.Find("[class*='ad'], [id*='ad'], [class*='advertisement']").Remove()

// Pass 3: Remove iframes and clutter
doc.Find("iframe, footer, aside").Remove()

// Pass 4: Remove elements with clutter in class names
doc.Find("[class*='sidebar'], [class*='menu'], [class*='popup']").Remove()
doc.Find("[class*='modal'], [class*='cookie']").Remove()

// Pass 5: Strip inline attributes (except semantic ones)
doc.Find("*").Each(func(i int, s *goquery.Selection) {
    // Keep only: href, src, alt, title
})
```

## CLI Interface

### Command-line Flags
```bash
webfetch-clean [FLAGS]

--mcp                   # Run as MCP server (JSON-RPC via stdio)
--url <url>             # URL to fetch (required in CLI mode)
--format <format>       # Output: html or markdown (default: markdown)
--preserve-main         # Only keep <main>/<article> content
--remove-images         # Remove all images
--timeout <seconds>     # HTTP timeout (default: 30)
--output <file>         # Write to file (default: stdout)
--help                  # Show help message
```

### Usage Examples
```bash
# Fetch and convert to markdown
webfetch-clean --url https://example.com

# Output cleaned HTML to file
webfetch-clean --url https://example.com --format html --output cleaned.html

# Only preserve main content
webfetch-clean --url https://example.com --preserve-main

# Run as MCP server
webfetch-clean --mcp
```

## Implementation Phases

### Phase 1: Project Setup ✓
- [x] Create project directory
- [ ] Initialize Go module: `go mod init github.com/hegner123/webfetch-clean`
- [ ] Add dependencies: `go get github.com/PuerkitoBio/goquery github.com/JohannesKaufmann/html-to-markdown`
- [ ] Create `.gitignore`

### Phase 2: Core HTTP Fetcher
**File:** `fetcher.go`
- Implement HTTP client with configurable timeout
- Set proper User-Agent header
- Handle redirects, 4xx/5xx errors
- Validate response content-type
- Return raw HTML string

**Function signature:**
```go
func FetchURL(url string, timeout int) (string, error)
```

### Phase 3: HTML Cleaning Logic
**File:** `cleaner.go`
- Parse HTML with goquery
- Implement multi-pass cleaning pipeline
- Add `preserveMainOnly` option
- Add `removeImages` option
- Return cleaned HTML string

**Function signature:**
```go
func CleanHTML(html string, preserveMainOnly bool, removeImages bool) (string, error)
```

### Phase 4: Format Conversion
**File:** `converter.go`
- Wrap html-to-markdown library
- Support both "html" and "markdown" output
- Handle conversion errors gracefully
- Preserve code blocks, links, lists, tables

**Function signature:**
```go
func ConvertToMarkdown(html string) (string, error)
```

### Phase 5: Main Entry Point
**File:** `main.go`
- Copy JSON-RPC 2.0 protocol from checkfor
- Implement dual-mode routing (CLI vs MCP)
- Add CLI flag parsing
- Implement MCP methods: `initialize`, `tools/list`, `tools/call`
- Wire up: fetcher → cleaner → converter pipeline
- Return results as JSON

### Phase 6: Configuration
**File:** `.mcp.json`
```json
{
  "mcpServers": {
    "webfetch-clean": {
      "command": "webfetch-clean",
      "args": ["--mcp"]
    }
  }
}
```

### Phase 7: Testing
- Test CLI mode with various URLs
- Test MCP mode with JSON-RPC requests
- Test both HTML and Markdown output
- Test edge cases: timeouts, 404s, malformed HTML
- Verify cleaning effectiveness

### Phase 8: Installation
1. Build binary: `go build -o webfetch-clean`
2. Install: `sudo cp webfetch-clean /usr/local/bin/`
3. Verify CLI works
4. Verify MCP integration with Claude Code

### Phase 9: Documentation
- Update `/Users/home/.claude/CLAUDE.md` with usage instructions
- Write comprehensive README.md
- Document all CLI flags
- Add usage examples

## Testing Strategy

### Manual Tests
```bash
# Test 1: CLI mode - Markdown output
webfetch-clean --url https://news.ycombinator.com --format markdown

# Test 2: CLI mode - HTML output
webfetch-clean --url https://example.com --format html

# Test 3: MCP mode - Initialize
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | webfetch-clean --mcp

# Test 4: MCP mode - Tools list
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | webfetch-clean --mcp

# Test 5: MCP mode - Tool call
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"webfetch_clean","arguments":{"url":"https://example.com"}}}' | webfetch-clean --mcp
```

### Success Criteria
- ✅ Binary builds without errors
- ✅ CLI mode fetches and cleans URLs successfully
- ✅ MCP mode responds to JSON-RPC correctly
- ✅ Both HTML and Markdown output work
- ✅ Cleaning removes scripts, styles, nav, ads, iframes
- ✅ Semantic content preserved (headings, paragraphs, lists, code)
- ✅ Tool appears in Claude Code's available tools
- ✅ Claude automatically uses tool when appropriate

## Error Handling

### HTTP Errors
- Network failures → "Failed to fetch URL: [error]"
- 4xx errors → "Page not found or forbidden (HTTP [code])"
- 5xx errors → "Server error (HTTP [code])"
- Timeout → "Request timeout after [N] seconds"

### Parsing Errors
- Invalid HTML → Use goquery's lenient parser (rarely fails)
- Empty response → "No content received from URL"

### Conversion Errors
- Markdown conversion failure → Return cleaned HTML as fallback
- Encoding issues → Use UTF-8 detection and conversion

### MCP Protocol Errors
- -32700: Parse error
- -32600: Invalid Request
- -32601: Method not found
- -32602: Invalid params
- -32603: Internal error

## Critical Reference Files
1. `/Users/home/Documents/Code/Go_dev/checkfor/main.go` - MCP protocol pattern
2. `/Users/home/Documents/Code/Go_dev/checkfor/.mcp.json` - Configuration pattern

## Future Enhancements
- Custom cleaning rules via config file
- Caching for repeated URLs
- Batch URL processing
- Readability scoring
- Metadata extraction (OpenGraph, Twitter Cards)
- PDF output support

## Development Notes
- Use goquery for ALL HTML manipulation (no regex on HTML)
- Standard library net/http is sufficient
- Follow Go best practices: early returns, minimal nesting
- Keep it simple - match checkfor's implementation style
- Test with real websites: news sites, documentation, blogs
