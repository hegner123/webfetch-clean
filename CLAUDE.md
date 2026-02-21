# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

webfetch-clean is a high-performance MCP tool that fetches web pages, removes clutter (ads, scripts, navigation), and outputs clean HTML or Markdown. It provides 90-96% token cost savings compared to Claude's built-in WebFetch tool by processing HTML locally rather than sending raw HTML through the API.

**Triple-Mode Operation:**
- **MCP Server Mode (default)**: JSON-RPC 2.0 protocol for Claude Code integration via stdin/stdout
- **CLI Mode**: Command-line tool with `--cli` flag
- **HTTP Mode**: REST API server with `--http` flag for remote access with API key auth

**Processing Modes:**
- **Clean mode (default)**: Aggressively strips ads, scripts, styles, nav, footer, sidebars, and non-semantic attributes for AI token efficiency
- **Scrape mode**: Light processing — only removes `<head>`, preserving page structure (scripts, styles, nav, footer, ads, iframes, attributes) for HTML ingestion workflows

## Development Commands

### Build and Install
```bash
# Build binary
go build -o webfetch-clean

# Install to /usr/local/bin (requires sudo)
sudo cp webfetch-clean /usr/local/bin/

# Verify installation
webfetch-clean --help
```

### Testing

**Run all tests:**
```bash
go test -v ./...
```

**Run specific test:**
```bash
go test -v -run TestCleanHTML_RemoveAds
```

**Run with coverage:**
```bash
go test -v -coverprofile="coverage.out" -covermode=atomic .
go tool cover -html="coverage.out"  # View HTML report
go tool cover -func="coverage.out"  # View function coverage
```

**Run with race detection:**
```bash
go test -v -race .
```

### Code Quality

**Format code:**
```bash
gofmt -w .
```

**Lint code:**
```bash
golangci-lint run
```

**Check for issues:**
```bash
go vet ./...
```

### MCP Testing

**Test MCP server mode:**
```bash
# Initialize
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | webfetch-clean

# List tools
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | webfetch-clean

# Call tool
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"webfetch_clean","arguments":{"url":"https://example.com"}}}' | webfetch-clean

# Call tool in scrape mode
echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"webfetch_clean","arguments":{"url":"https://example.com","mode":"scrape"}}}' | webfetch-clean
```

### CLI Testing

**Test CLI mode with URL:**
```bash
# Basic fetch (markdown output)
webfetch-clean --cli --url https://example.com

# HTML output
webfetch-clean --cli --url https://example.com --format html

# Save to file
webfetch-clean --cli --url https://example.com --output result.md

# Preserve only main content
webfetch-clean --cli --url https://example.com --preserve-main

# Remove images
webfetch-clean --cli --url https://example.com --remove-images

# Custom timeout
webfetch-clean --cli --url https://example.com --timeout 60

# Scrape mode — preserves page structure (scripts, styles, nav, footer, etc.)
webfetch-clean --cli --url https://example.com --mode scrape

# Scrape mode with HTML output
webfetch-clean --cli --url https://example.com --mode scrape --format html
```

**Test CLI mode with local file:**
```bash
# Basic file processing (markdown output)
webfetch-clean --cli --file test.html

# HTML output from file
webfetch-clean --cli --file test.html --format html

# Save file output to another file
webfetch-clean --cli --file input.html --output result.md

# Preserve only main content from file
webfetch-clean --cli --file test.html --preserve-main

# Remove images from file
webfetch-clean --cli --file test.html --remove-images

# All options work with file input
webfetch-clean --cli --file test.html --preserve-main --remove-images --format html

# Custom output limit
webfetch-clean --cli --file test.html --max-tokens 50000
```

### HTTP Server Testing

**Start HTTP server:**
```bash
webfetch-clean --http :8080 --api-key my-secret --base-url http://localhost:8080
```

**Test endpoints:**
```bash
# Health check (no auth)
curl localhost:8080/health

# Initialize
curl -H "X-API-Key: my-secret" -X POST localhost:8080/mcp \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize"}'

# List tools (HTTP schema with file_token, result_id, override)
curl -H "X-API-Key: my-secret" -X POST localhost:8080/mcp \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list"}'

# Fetch URL
curl -H "X-API-Key: my-secret" -X POST localhost:8080/mcp \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"webfetch_clean","arguments":{"url":"https://example.com"}}}'

# Create file token
curl -H "X-API-Key: my-secret" -X POST localhost:8080/admin/tokens \
  -d '{"file":"/path/to/file.html","expires_minutes":60}'

# Use file token
curl -H "X-API-Key: my-secret" -X POST localhost:8080/mcp \
  -d '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"webfetch_clean","arguments":{"file_token":"TOKEN_HERE"}}}'

# Download over-limit result
curl -H "X-API-Key: my-secret" localhost:8080/results/RESULT_ID
```

**HTTP server flags:**
- `--http :8080` — Bind address (default loopback; use `0.0.0.0:8080` for all interfaces)
- `--api-key SECRET` — API key (or `WEBFETCH_API_KEY` env var; required)
- `--base-url URL` — Public base URL for download links in over-limit responses
- `--db webfetch.db` — SQLite database path for file tokens

## Architecture

The codebase follows a clean pipeline architecture with three core processing stages:

```
Fetch/Read → Clean → Convert
```

### Input Sources

The tool supports two input sources in CLI mode:
- **URL mode**: Fetches HTML from a remote URL (uses `FetchURL`)
- **File mode**: Reads HTML from a local file (uses `ReadFile`)

Both sources feed into the same cleaning and conversion pipeline.

### Core Components

**main.go** - Entry point, MCP protocol handler, mode routing
- Routes to HTTP, CLI, or stdio MCP mode based on flags
- Implements JSON-RPC 2.0 protocol for stdio MCP mode
- Handles `initialize`, `tools/list`, `tools/call` methods
- CLI flag parsing and execution
- `ReadFile` function for local file input
- `processInput()` — pure pipeline function (no I/O at limit-check step)

**httpserver.go** - HTTP server with API key auth
- `HTTPServer` struct with routes, TempStore, TokenStore
- `TempStore` — bounded in-memory store for oversized results (100 entries, 60s TTL)
- Routes: `POST /mcp`, `GET /results/{id}`, `POST /admin/tokens`, `GET /health`
- Auth middleware: `X-API-Key` header with constant-time comparison
- HTTP tool schema: `file_token` and `result_id`/`override` instead of raw `file`
- Over-limit: stores in TempStore, returns structured `OverLimitResult` with retrieval options
- Graceful shutdown on SIGINT/SIGTERM

**tokenstore.go** - SQLite-backed file access tokens
- `TokenStore` wrapping sqlc-generated queries
- `CreateFileToken(filePath, ttl)` — validates file, generates UUID, stores with expiry
- `RedeemToken(token)` — atomic single-statement redemption (no TOCTOU race)
- `Cleanup()` — deletes expired and consumed rows
- `generateUUID()` — RFC 4122 v4 using crypto/rand

**fetcher.go** - HTTP client with timeout and error handling
- `FetchURL(url, timeout)` - Fetches HTML content
- HTTP headers: User-Agent, Accept, Accept-Language
- Status code validation (400s, 500s)
- Error wrapping with context

**cleaner.go** - Multi-pass HTML cleaning pipeline
- `CleanHTML(html, preserveMainOnly, removeImages, stripLinks, mode)` - Main cleaning function
- `mode` parameter: `"clean"` (default) or `"scrape"`
- **Clean mode** runs all passes:
  - Pass 1: Remove `<head>`, `<script>`, `<style>`, `<nav>`
  - Pass 2: Remove ad-related elements (class/id patterns)
  - Pass 3: Remove tracking iframes
  - Pass 4: Remove clutter (footer, aside, sidebar, menu, popup, modal, cookie, social, comments)
  - Pass 5: Strip inline attributes (keeps only href, src, alt, title)
- **Scrape mode** only removes `<head>`, preserves everything else
- `removeImages`, `stripLinks`, and `preserveMainOnly` apply in both modes
- Preserves semantic elements: `<main>`, `<article>`, `<p>`, `<h1-h6>`, `<ul>`, `<ol>`, `<code>`, `<pre>`, `<table>`, `<a>`, `<img>`, `<blockquote>`

**converter.go** - HTML-to-Markdown conversion
- `ConvertToFormat(html, format)` - Routes to HTML or Markdown
- `ConvertToMarkdown(html)` - Uses html-to-markdown library

### Dependencies

```go
// go.mod (Go 1.25.5+)
github.com/PuerkitoBio/goquery v1.11.0        // jQuery-like HTML parsing
github.com/JohannesKaufmann/html-to-markdown v1.6.0  // HTML to Markdown conversion
modernc.org/sqlite v1.46.1                   // Pure Go SQLite driver (no CGO)
```

### Data Flow

**MCP Mode:**
1. Claude Code sends JSON-RPC request to stdin
2. main.go parses request and routes to handler
3. For `tools/call`: FetchURL → CleanHTML → ConvertToFormat
4. Result returned as JSON-RPC response to stdout

**CLI Mode:**
1. Parse CLI flags (`--cli`, `--url` or `--file`, `--format`, etc.)
2. FetchURL (if `--url`) OR ReadFile (if `--file`) → CleanHTML → ConvertToFormat
3. If over limit: write to file. Otherwise: write output to file or stdout

**HTTP Mode:**
1. HTTP request hits `/mcp` with JSON-RPC body and `X-API-Key` header
2. Route to method handler (initialize, tools/list, tools/call)
3. For `tools/call`: resolve file_token or use URL → processInput pipeline
4. If over limit: store in TempStore, return `OverLimitResult` with result_id
5. Client can retrieve via `result_id`+`override` or `GET /results/{id}`

## Code Style

### Go-Specific Patterns

**Use `any` not `interface{}`:**
```go
// Good
type JSONRPCRequest struct {
    ID any `json:"id"`
}

// Bad
type JSONRPCRequest struct {
    ID interface{} `json:"id"`
}
```

**Error handling with wrapping:**
```go
// Always wrap errors with context
if err != nil {
    return "", fmt.Errorf("failed to fetch URL: %w", err)
}
```

**Early returns over nesting:**
```go
// Good
func ProcessURL(url string) error {
    if url == "" {
        return fmt.Errorf("URL cannot be empty")
    }

    html, err := FetchURL(url, 30)
    if err != nil {
        return fmt.Errorf("fetch failed: %w", err)
    }

    return nil
}

// Bad
func ProcessURL(url string) error {
    if url != "" {
        html, err := FetchURL(url, 30)
        if err == nil {
            return nil
        } else {
            return err
        }
    }
    return fmt.Errorf("URL cannot be empty")
}
```

### Testing Patterns

**Table-driven tests:**
```go
func TestCleanHTML(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "removes script tags",
            input:   "<html><script>alert('hi')</script><p>content</p></html>",
            want:    "<html><p>content</p></html>",
            wantErr: false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := CleanHTML(tt.input, false, false, false, "clean")
            if (err != nil) != tt.wantErr {
                t.Errorf("CleanHTML() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("CleanHTML() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Output Limits

### Token-Based Size Management

webfetch-clean implements automatic output size management to prevent excessive token consumption:

**Default Limit:** 100,000 tokens (300KB, calculated as 3 bytes = 1 token)

**Behavior:**
- If output is under the limit: Returns content normally
- If output exceeds the limit: Writes content to a file and returns a message with the filename

**Generated Filenames:**
- URL input: `https://example.com/foo/bar` → `example-com-foo-bar.md`
- File input: `/path/to/file.html` → `file-cleaned.md`

**Configuration:**
- CLI: `--max-tokens 50000` (set custom limit)
- MCP: `max_tokens` parameter (default: 100000)

**Example:**
```bash
# Use default 100k token limit
webfetch-clean --cli --url https://example.com

# Set custom limit of 50k tokens
webfetch-clean --cli --url https://example.com --max-tokens 50000
```

When limit is exceeded, you'll see:
```
Output exceeded limit (120000 tokens, limit: 100000 tokens). Content written to file: example-com.md
```

## Known Limitations

### Ad Detection Trade-offs

The ad detection is intentionally aggressive. It matches patterns like:
- `advertisement`, `ad-`, `-ad-`, `-ad`, `_ad_`, `_ad`, `ad_`, ` ad `

This means legitimate content with "ad" in class names may be removed:
- `thread-card` matches `-ad-` pattern
- `reader-mode` matches `-ad-` pattern

This is acceptable because:
1. Better to be overly aggressive in removing ads
2. Most legitimate content doesn't have "ad" surrounded by dashes
3. Main semantic content is usually not affected
4. Scrape mode (`--mode scrape`) bypasses all ad detection, preserving everything

## CI/CD

**GitHub Actions workflow:** `.github/workflows/test.yml`

**Jobs:**
- **test**: Run tests on Ubuntu/macOS with Go 1.21/1.22/1.23, race detection, coverage upload to Codecov
- **lint**: Run golangci-lint with config from `.golangci.yml`
- **build**: Build binary on multiple platforms, test execution, upload artifacts

**Linters enabled:**
- errcheck, gosimple, govet, ineffassign, staticcheck, unused
- gofmt, goimports, misspell, revive

## Development Workflow

**Branch naming:**
- `feature/*` - New features
- `fix/*` - Bug fixes
- `docs/*` - Documentation updates
- `refactor/*` - Code refactoring

**Commit message format (Conventional Commits):**
```
<type>(<scope>): <subject>

<body>

<footer>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`

**Examples:**
```
feat(cleaner): add support for removing cookie banners

Implement detection and removal of common cookie consent
banners by checking for class names containing 'cookie',
'consent', and 'gdpr'.

Closes #123
```

## Important Files

- **START.md**: Onboarding guide with implementation phases and progress tracking
- **TESTING.md**: Comprehensive testing guide with coverage reports and CI/CD details
- **CONTRIBUTING.md**: Full contribution guidelines with code standards and PR process
- **README.md**: User-facing documentation with installation and usage instructions
- **docs/CASE_STUDY.md**: Token cost analysis and savings calculations
- **docs/HTTP_SERVER.md**: HTTP server mode design spec and implementation plan
- **httpserver.go**: HTTP server, routes, TempStore, auth middleware
- **tokenstore.go**: SQLite-backed file access token management
- **sqlc/**: SQL schema and queries for sqlc code generation
- **db/**: sqlc-generated Go code (committed, no build-time sqlc needed)
