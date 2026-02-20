# webfetch-clean

A high-performance MCP tool that fetches web pages, removes clutter (ads, scripts, navigation), and outputs clean HTML or Markdown. Provides **90-96% token cost savings** compared to Claude's built-in WebFetch tool.

## Why Use This Tool?

When using Claude's built-in WebFetch tool, you pay for the **entire raw HTML** as input tokens (2,500-25,000+ tokens per page). webfetch-clean performs all processing locally, so you only pay for the cleaned output.

**Token Savings:**
- Simple page (10KB): Save ~2,334 tokens (93% reduction)
- Documentation (100KB): Save ~23,987 tokens (96% reduction)
- For detailed cost analysis, see [docs/CASE_STUDY.md](docs/CASE_STUDY.md)

## Features

- **Triple-Mode Operation:** Works as CLI tool, stdio MCP server, or HTTP server
- **Multi-Pass Cleaning:** Removes ads, scripts, styles, navigation, sidebars, popups, modals, social widgets, and comments
- **Format Options:** Output as HTML or Markdown
- **Content Preservation:** Keeps semantic content (headings, paragraphs, lists, code blocks, tables, links)
- **Zero API Tokens:** All processing happens locally in compiled Go binary
- **MCP Protocol:** JSON-RPC 2.0 compatible for Claude Code integration

## Installation

### Download from Release

Download a pre-built binary from [GitHub Releases](https://github.com/hegner123/webfetch-clean/releases/latest):

**macOS:**
```bash
# Apple Silicon (M1/M2/M3/M4)
curl -L https://github.com/hegner123/webfetch-clean/releases/latest/download/webfetch-clean-darwin-arm64 -o webfetch-clean

# Intel
curl -L https://github.com/hegner123/webfetch-clean/releases/latest/download/webfetch-clean-darwin-amd64 -o webfetch-clean

chmod +x webfetch-clean
sudo mv webfetch-clean /usr/local/bin/
```

**Linux:**
```bash
# x86_64
curl -L https://github.com/hegner123/webfetch-clean/releases/latest/download/webfetch-clean-linux-amd64 -o webfetch-clean

# ARM64
curl -L https://github.com/hegner123/webfetch-clean/releases/latest/download/webfetch-clean-linux-arm64 -o webfetch-clean

chmod +x webfetch-clean
sudo mv webfetch-clean /usr/local/bin/
```

**Windows (PowerShell):**
```powershell
# x86_64
Invoke-WebRequest -Uri https://github.com/hegner123/webfetch-clean/releases/latest/download/webfetch-clean-windows-amd64.exe -OutFile webfetch-clean.exe

# ARM64
Invoke-WebRequest -Uri https://github.com/hegner123/webfetch-clean/releases/latest/download/webfetch-clean-windows-arm64.exe -OutFile webfetch-clean.exe
```

### Build from Source

Requires Go 1.23 or later.

```bash
git clone https://github.com/hegner123/webfetch-clean.git
cd webfetch-clean
go build -o webfetch-clean
sudo cp webfetch-clean /usr/local/bin/
```

### Add to Claude Code

After building, register the MCP server with Claude Code:

```bash
# If installed to /usr/local/bin (in PATH)
claude mcp add --scope user --transport stdio webfetch-clean -- webfetch-clean

# If using absolute path (not in PATH)
claude mcp add --scope user --transport stdio webfetch-clean -- $HOME/path/to/webfetch-clean
```

Verify it's registered:

```bash
claude mcp list
```

Then inside a Claude Code session, run `/mcp` to confirm the server is connected.

### Verify Installation

```bash
webfetch-clean --cli --url https://example.com
```

### Optional Install Script

For convenience, you can create your own `install.sh`:

```bash
#!/usr/bin/env bash
set -e

echo "Building webfetch-clean..."
go build -o webfetch-clean

if [ ! -f "webfetch-clean" ]; then
    echo "Error: Build failed. webfetch-clean binary not found."
    exit 1
fi

echo ""
echo "Build successful!"
echo ""
echo "Installing to /usr/local/bin (requires sudo)..."
sudo cp webfetch-clean /usr/local/bin/

if [ $? -eq 0 ]; then
    echo "Success! webfetch-clean installed to /usr/local/bin/"
    echo ""
    echo "Verify installation:"
    echo "  webfetch-clean --help"
else
    echo "Error: Installation failed."
    exit 1
fi
```

Then make it executable and run: `chmod +x install.sh && ./install.sh`

## Usage

### MCP Server Mode (Default)

By default, `webfetch-clean` runs as an MCP server for Claude Code integration:

```bash
webfetch-clean
```

#### Adding to Claude Code

**When webfetch-clean is in PATH (installed to /usr/local/bin):**
```bash
# Command name only - relies on PATH
claude mcp add --scope user --transport stdio webfetch-clean -- webfetch-clean
```

**When webfetch-clean is in a custom directory (not in PATH):**
```bash
# Use absolute path with $HOME (not ~)

# Example: binary in Downloads
claude mcp add --scope user --transport stdio webfetch-clean -- $HOME/Downloads/webfetch-clean-binary/webfetch-clean

# Example: binary in project directory
claude mcp add --scope user --transport stdio webfetch-clean -- $HOME/Documents/Code/Go_dev/webfetch-clean/webfetch-clean

# Example: binary in custom tools directory
claude mcp add --scope user --transport stdio webfetch-clean -- $HOME/bin/webfetch-clean
```

**Key points:**
- Use `$HOME` instead of `~` for absolute paths
- The `--` separates the server name from the command/args
- User scope (`--scope user`) makes it available across all projects
- Both scenarios work identically once configured

**Verify configuration:**
```bash
claude mcp get webfetch-clean
claude mcp list
```

The MCP server mode uses JSON-RPC 2.0 protocol and is automatically available in Claude Code once configured.

#### Update Global CLAUDE.md

Add usage instructions to your global `~/.claude/CLAUDE.md` so Claude knows when to use this tool:

```markdown
## Tool Usage - Search Optimization

### When to use webfetch_clean (MCP tool)

**CRITICAL: ALWAYS use webfetch_clean instead of the built-in WebFetch tool for ALL web fetching tasks.**

The `webfetch_clean` tool is the preferred default for fetching web pages. It provides cleaner output with ads, scripts, navigation, and clutter removed.

**Use webfetch_clean as the default for:**
- ALL web page fetching (documentation, blogs, articles, news sites)
- Any URL that returns HTML content
- Fetching documentation sites with heavy navigation
- Fetching blog posts or articles with ads and sidebars
- Analyzing web pages where clutter interferes with understanding
- General web content retrieval (unless explicitly told otherwise)

**Parameters:**
```
webfetch_clean tool with:
- url: "https://example.com" (required)
- output_format: "markdown" or "html" (default: "markdown")
- preserve_main_only: false (default: false, set true to extract only main/article content)
- remove_images: false (default: false, set true to remove all images)
- timeout: 30 (default: 30 seconds)
```

**What it removes:**
- `<head>`, `<script>`, `<style>`, `<nav>` elements
- Ad-related elements (class/id containing: ad, advertisement, banner)
- Tracking iframes
- Clutter (footer, aside, sidebar, menu, popup, modal, cookie, social, share, comments)
- Inline attributes (keeps only href, src, alt, title)

**What it preserves:**
- Main semantic content (main, article, p, h1-h6, ul, ol, code, pre, table, a, img)

**ONLY use the built-in WebFetch tool as a fallback when:**
- webfetch_clean fails or returns an error
- webfetch_clean is unavailable (MCP server down)
- Simple API endpoints returning JSON/XML (not HTML)
- Plain text pages without HTML markup
- User explicitly requests "use WebFetch" or "raw HTML"
- You specifically need the unprocessed, raw HTML with all scripts/styles intact

**Default behavior: Always try webfetch_clean first. Only fall back to WebFetch if webfetch_clean fails.**
```

This instructs Claude to automatically use webfetch_clean for web content retrieval.

### HTTP Server Mode

HTTP mode exposes the same MCP JSON-RPC interface over HTTP with API key authentication. This allows remote team members and AI orchestrators to use webfetch-clean without installing it locally.

```bash
webfetch-clean --http :8080 --api-key my-secret --base-url http://localhost:8080
```

#### HTTP Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--http` | string | | Bind address (e.g., `:8080`). Defaults to `127.0.0.1` if only port given |
| `--api-key` | string | | API key for auth (or `WEBFETCH_API_KEY` env var). Required |
| `--base-url` | string | | Public URL for download links (e.g., `https://fetch.example.com`) |
| `--db` | string | `webfetch.db` | SQLite database path for file access tokens |

#### Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/mcp` | Yes | JSON-RPC 2.0 handler (same protocol as stdio mode) |
| `GET` | `/results/{id}` | Yes | Download oversized content by result ID |
| `POST` | `/admin/tokens` | Yes | Create a single-use file access token |
| `GET` | `/health` | No | Health check |

All authenticated endpoints require the `X-API-Key` header.

#### Adding to Claude Code (HTTP transport)

Register the remote server with Claude Code using HTTP transport:

```bash
# Local server
claude mcp add --transport http webfetch-clean http://localhost:8080/mcp

# Remote server behind a reverse proxy
claude mcp add --transport http webfetch-clean https://fetch.example.com/mcp
```

#### Fetching URLs

Send a standard MCP `tools/call` request over HTTP:

```bash
curl -H "X-API-Key: my-secret" -X POST http://localhost:8080/mcp \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"webfetch_clean","arguments":{"url":"https://example.com"}}}'
```

The HTTP tool schema includes the same parameters as stdio mode (`url`, `output_format`, `mode`, `preserve_main_only`, `remove_images`, `strip_links`, `timeout`, `max_tokens`) plus these HTTP-specific parameters:

| Parameter | Type | Description |
|-----------|------|-------------|
| `file_token` | string | Single-use token for server-side file access (replaces `file`) |
| `result_id` | string | ID of a stored over-limit result |
| `override` | boolean | When `true` with `result_id`, returns full stored content |

The raw `file` parameter is not available in HTTP mode. Use file tokens instead.

#### File Tokens

File tokens provide secure, single-use access to server-side files without exposing raw paths over HTTP.

**Create a token:**
```bash
curl -H "X-API-Key: my-secret" -X POST http://localhost:8080/admin/tokens \
  -d '{"file":"/path/to/file.html","expires_minutes":60}'
```

Response:
```json
{"token": "550e8400-e29b-41d4-a716-446655440000", "expires": "2026-01-01T01:00:00Z"}
```

**Use the token:**
```bash
curl -H "X-API-Key: my-secret" -X POST http://localhost:8080/mcp \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"webfetch_clean","arguments":{"file_token":"550e8400-e29b-41d4-a716-446655440000"}}}'
```

Tokens are single-use (consumed on first redemption) and expire after the specified duration (default: 60 minutes, max: 24 hours). Expired and consumed tokens are cleaned up automatically.

#### Over-Limit Results

When output exceeds the token limit (default: 100,000 tokens), the HTTP server stores the content temporarily instead of writing to a local file. The response includes retrieval options:

```json
{
  "result_id": "uuid-here",
  "token_count": 120000,
  "limit": 100000,
  "message": "Fetched https://example.com — 120000 tokens (limit: 100000).\nResult stored for 60 seconds (ID: uuid-here).\n\nOptions:\n1. Override limit: call webfetch_clean with {\"result_id\": \"uuid-here\", \"override\": true}\n2. Download: GET http://localhost:8080/results/uuid-here\n3. Do nothing — result expires in 60 seconds."
}
```

**Retrieve via override:**
```bash
curl -H "X-API-Key: my-secret" -X POST http://localhost:8080/mcp \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"webfetch_clean","arguments":{"result_id":"uuid-here","override":true}}}'
```

**Retrieve via download:**
```bash
curl -H "X-API-Key: my-secret" http://localhost:8080/results/uuid-here
```

Results are stored in memory (max 100 entries) and expire after 60 seconds.

#### Docker Deployment

The project includes a Dockerfile and Docker Compose configuration with a Caddy reverse proxy for production deployment.

**Quick start:**
```bash
# Set required environment variables
export WEBFETCH_API_KEY=your-secret-key
export BASE_URL=https://fetch.example.com
export SITE_ADDRESS=fetch.example.com

# Start services
docker compose up -d
```

**Docker Compose services:**

| Service | Description |
|---------|-------------|
| `webfetch` | webfetch-clean HTTP server on port 8080 (internal) |
| `caddy` | Reverse proxy with automatic TLS on ports 80/443 |

**Environment variables:**

| Variable | Required | Description |
|----------|----------|-------------|
| `WEBFETCH_API_KEY` | Yes | API key for authentication |
| `BASE_URL` | Yes | Public URL for download links |
| `SITE_ADDRESS` | No | Domain for Caddy TLS (default: `localhost`) |

The Caddy proxy automatically provisions TLS certificates when `SITE_ADDRESS` is set to a public domain. For local development, the default `localhost` uses Caddy's internal CA.

Data is persisted in Docker volumes (`webfetch-data` for the SQLite database, `caddy-data` and `caddy-config` for Caddy state).

### CLI Mode

Use the `--cli` flag for command-line usage:

```bash
# Fetch and convert to markdown (default)
webfetch-clean --cli --url https://example.com

# Output as HTML
webfetch-clean --cli --url https://example.com --format html

# Save to file
webfetch-clean --cli --url https://example.com --output result.md

# Only preserve main content
webfetch-clean --cli --url https://example.com --preserve-main

# Remove images
webfetch-clean --cli --url https://example.com --remove-images

# Custom timeout (default: 30s)
webfetch-clean --cli --url https://example.com --timeout 60
```

### CLI Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--cli` | bool | `false` | Run in CLI mode (default: MCP server mode) |
| `--url` | string | | URL to fetch (alternative to `--file`) |
| `--file` | string | | Local HTML file to process (alternative to `--url`) |
| `--format` | string | `markdown` | Output format: `html` or `markdown` |
| `--mode` | string | `clean` | Processing mode: `clean` or `scrape` |
| `--preserve-main` | bool | `false` | Only preserve `<main>`/`<article>` content |
| `--remove-images` | bool | `false` | Remove all images from output |
| `--strip-links` | bool | `false` | Replace links with their text content |
| `--browser` | bool | `false` | Use headless browser for JavaScript-rendered pages |
| `--timeout` | int | `30` | HTTP request timeout in seconds |
| `--output` | string | stdout | Write output to file instead of stdout |
| `--verbose` | bool | `false` | Print progress messages to stderr |

## What It Removes

- `<head>`, `<script>`, `<style>`, `<nav>` elements
- Ad-related elements (class/id containing: ad, advertisement, banner)
- Tracking iframes
- Clutter: footer, aside, sidebar, menu, popup, modal, cookie banners
- Social media widgets and share buttons
- Comment sections
- All inline attributes (except href, src, alt, title)

## What It Preserves

- Semantic HTML: `<main>`, `<article>`, `<p>`, `<h1-h6>`, `<ul>`, `<ol>`, `<li>`
- Code blocks: `<code>`, `<pre>`
- Tables: `<table>`, `<tr>`, `<td>`, `<th>`
- Links and images: `<a>`, `<img>`
- Blockquotes: `<blockquote>`

## MCP Tool Schema

When used with Claude Code, the tool is available as `webfetch_clean`.

**Stdio mode parameters:**

```json
{
  "name": "webfetch_clean",
  "parameters": {
    "url": "string — URL to fetch",
    "file": "string — local file path (alternative to url)",
    "output_format": "html | markdown (default: markdown)",
    "mode": "clean | scrape (default: clean)",
    "preserve_main_only": "boolean (default: false)",
    "remove_images": "boolean (default: false)",
    "strip_links": "boolean (default: false)",
    "timeout": "integer (default: 30)",
    "max_tokens": "integer (default: 100000)"
  }
}
```

**HTTP mode parameters** (same as above, plus):

```json
{
  "file_token": "string — single-use token for server-side file access (replaces file)",
  "result_id": "string — ID of a stored over-limit result",
  "override": "boolean — retrieve full content for a result_id (default: false)"
}
```

The `file` and `output_directory` parameters are not available in HTTP mode.

## Integration with Claude Code

Add to your `~/.claude/CLAUDE.md`:

```markdown
### When to use webfetch_clean (MCP tool)
Use `webfetch_clean` instead of WebFetch for 90-96% token cost savings.

**Use webfetch_clean when:**
- Fetching documentation, blog posts, or articles
- You need complete, accurate content (not AI summaries)
- Token efficiency matters (saves 2,500-25,000+ tokens per page)

**Parameters:**
- url: "https://example.com" (required)
- output_format: "markdown" or "html" (default: "markdown")
- preserve_main_only: false (default, set true for main/article only)
- remove_images: false (default, set true to remove all images)
- timeout: 30 (default timeout in seconds)
```

## Architecture

```
webfetch-clean/
├── main.go           # Entry point, MCP protocol, CLI routing
├── httpserver.go     # HTTP server, routes, TempStore, auth middleware
├── tokenstore.go     # SQLite-backed file access token management
├── fetcher.go        # HTTP client with timeout/error handling
├── cleaner.go        # Multi-pass HTML cleaning pipeline
├── converter.go      # HTML-to-Markdown conversion
├── db/               # sqlc-generated Go code (committed)
├── sqlc/             # SQL schema and queries for sqlc
├── Dockerfile        # Multi-stage build (Go build + Debian runtime with Chromium)
├── docker-compose.yml # webfetch + Caddy reverse proxy
├── Caddyfile         # Reverse proxy with automatic TLS
├── go.mod            # Dependencies
├── .mcp.json         # MCP server configuration
├── docs/
│   └── CASE_STUDY.md # Token cost analysis
└── README.md         # This file
```

### Dependencies

- [goquery](https://github.com/PuerkitoBio/goquery) - jQuery-like HTML parsing
- [html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) - HTML to Markdown conversion
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) - Pure Go SQLite driver (no CGO)

## Testing

### Test CLI Mode

```bash
# Simple page
webfetch-clean --cli --url https://example.com

# Documentation
webfetch-clean --cli --url https://go.dev/doc/effective_go

# News site
webfetch-clean --cli --url https://news.ycombinator.com
```

### Test MCP Mode

```bash
# Initialize
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | webfetch-clean

# List tools
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | webfetch-clean

# Call tool
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"webfetch_clean","arguments":{"url":"https://example.com"}}}' | webfetch-clean
```

### Test HTTP Mode

```bash
# Start server
webfetch-clean --http :8080 --api-key test123 --base-url http://localhost:8080

# Health check (no auth required)
curl http://localhost:8080/health

# Initialize
curl -H "X-API-Key: test123" -X POST http://localhost:8080/mcp \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize"}'

# List tools
curl -H "X-API-Key: test123" -X POST http://localhost:8080/mcp \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list"}'

# Fetch a URL
curl -H "X-API-Key: test123" -X POST http://localhost:8080/mcp \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"webfetch_clean","arguments":{"url":"https://example.com"}}}'
```

## Performance

- **Token Efficiency:** 90-96% reduction vs Claude WebFetch
- **Speed:** Local processing (no API latency)
- **Memory:** Minimal (~10-20MB for typical pages)
- **No Rate Limits:** Local execution means no API throttling

## Cost Analysis

For a developer fetching 100 documentation pages per day:

| Tool | Tokens/Day | Monthly Tokens | Monthly Cost |
|------|------------|----------------|--------------|
| **Claude WebFetch** | 2,500,000 | 75,000,000 | ~$450/month |
| **webfetch-clean** | 100,000 | 3,000,000 | ~$18/month |
| **Savings** | 2,400,000 | 72,000,000 | ~$432/month |

**Annual Savings: ~$5,184**

See [docs/CASE_STUDY.md](docs/CASE_STUDY.md) for detailed analysis with citations from Anthropic's documentation.

## When to Use Each Tool

### Use webfetch-clean (Recommended)
- Almost always - dramatically cheaper and more accurate
- Documentation research
- Content extraction
- Web scraping workflows
- When you want complete, accurate content

### Use Claude WebFetch
- When you explicitly want an AI summary instead of full content
- When you're willing to pay 10-30x more for that summary
- When processing content types not supported by webfetch-clean

## Error Handling

The tool provides clear error messages for common issues:

- **Network failures:** "Failed to fetch URL: [error]"
- **HTTP 4xx:** "Page not found or forbidden (HTTP [code])"
- **HTTP 5xx:** "Server error (HTTP [code])"
- **Timeouts:** "Request timeout after [N] seconds"
- **Empty response:** "No content received from URL"

## Contributing

Contributions are welcome! We appreciate bug reports, feature requests, documentation improvements, and code contributions.

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines on:
- Development workflow and setup
- Coding standards and style guide
- Testing requirements
- Commit message conventions
- Pull request process

Quick start:
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes following our coding standards
4. Add tests for new functionality
5. Commit using conventional commit format
6. Push and create a pull request

For bug reports and feature requests, please use our [issue templates](https://github.com/hegner123/webfetch-clean/issues/new/choose).

## License

MIT License - see LICENSE file for details

## Acknowledgments

- Built with [goquery](https://github.com/PuerkitoBio/goquery) by Martin Angers
- Uses [html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) by Johannes Kaufmann
- Follows MCP protocol specification from Anthropic
- Inspired by the need for cost-effective web content retrieval

## Support

- **Issues:** [GitHub Issues](https://github.com/hegner123/webfetch-clean/issues)
- **Discussions:** [GitHub Discussions](https://github.com/hegner123/webfetch-clean/discussions)

## Version

**Current Version:** 1.0.0

**Last Updated:** February 20, 2026

---

**Note:** Token cost estimates based on Anthropic's official documentation as of January 2026. See [docs/CASE_STUDY.md](docs/CASE_STUDY.md) for detailed analysis and sources.
