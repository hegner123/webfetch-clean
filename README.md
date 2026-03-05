# webfetch-clean

A high-performance MCP tool that fetches web pages, removes clutter (ads, scripts, navigation), and outputs clean HTML or Markdown. Provides **90-96% token cost savings** compared to sending raw HTML through AI APIs.

## Features

- **Triple-mode operation:** MCP server (stdio), CLI tool, or HTTP server
- **Two processing modes:** `clean` (aggressive clutter removal) and `scrape` (light processing, preserves structure)
- **Multi-pass cleaning:** Removes ads, scripts, styles, navigation, sidebars, popups, modals, social widgets, cookie banners, and comments
- **Format output:** HTML or Markdown
- **Headless browser support:** Render JavaScript-heavy pages with `--browser` flag
- **Token-aware:** Automatic output size management with configurable limits

## Installation

### Pre-built Binaries

Download from [GitHub Releases](https://github.com/hegner123/webfetch-clean/releases/latest):

```bash
# macOS Apple Silicon
curl -L https://github.com/hegner123/webfetch-clean/releases/latest/download/webfetch-clean-darwin-arm64 -o webfetch-clean

# macOS Intel
curl -L https://github.com/hegner123/webfetch-clean/releases/latest/download/webfetch-clean-darwin-amd64 -o webfetch-clean

# Linux x86_64
curl -L https://github.com/hegner123/webfetch-clean/releases/latest/download/webfetch-clean-linux-amd64 -o webfetch-clean

# Linux ARM64
curl -L https://github.com/hegner123/webfetch-clean/releases/latest/download/webfetch-clean-linux-arm64 -o webfetch-clean

chmod +x webfetch-clean
sudo mv webfetch-clean /usr/local/bin/
```

Windows binaries (`webfetch-clean-windows-amd64.exe`, `webfetch-clean-windows-arm64.exe`) are also available on the releases page.

### Build from Source

Requires Go 1.25+.

```bash
git clone https://github.com/hegner123/webfetch-clean.git
cd webfetch-clean
go build -o webfetch-clean
sudo cp webfetch-clean /usr/local/bin/
```

## Usage

### CLI Mode

```bash
# Fetch and convert to markdown
webfetch-clean --cli --url https://example.com

# Output as HTML
webfetch-clean --cli --url https://example.com --format html

# Process a local file
webfetch-clean --cli --file page.html

# Scrape mode (preserves page structure)
webfetch-clean --cli --url https://example.com --mode scrape

# Save to file
webfetch-clean --cli --url https://example.com --output result.md

# Only main/article content, no images
webfetch-clean --cli --url https://example.com --preserve-main --remove-images
```

#### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--cli` | `false` | Run in CLI mode |
| `--url` | | URL to fetch |
| `--file` | | Local HTML file to process |
| `--format` | `markdown` | Output format: `html` or `markdown` |
| `--mode` | `clean` | Processing mode: `clean` or `scrape` |
| `--preserve-main` | `false` | Only preserve `<main>`/`<article>` content |
| `--remove-images` | `false` | Remove all images |
| `--strip-links` | `false` | Replace links with text content |
| `--browser` | `false` | Use headless browser for JS-rendered pages |
| `--timeout` | `30` | HTTP timeout in seconds |
| `--max-tokens` | `100000` | Output size limit (3 bytes = 1 token) |
| `--output` | stdout | Write output to file |
| `--verbose` | `false` | Print progress to stderr |

### MCP Server Mode

The default mode. Register with Claude Code:

```bash
claude mcp add --scope user --transport stdio webfetch-clean -- webfetch-clean
```

Verify:

```bash
claude mcp list
```

MCP parameters: `url`, `file`, `output_format`, `mode`, `preserve_main_only`, `remove_images`, `strip_links`, `timeout`, `max_tokens`.

### HTTP Server Mode

Exposes the MCP interface over HTTP with API key authentication.

```bash
webfetch-clean --http :8080 --api-key my-secret --base-url http://localhost:8080
```

| Flag | Default | Description |
|------|---------|-------------|
| `--http` | | Bind address (e.g., `:8080`) |
| `--api-key` | | API key (or `WEBFETCH_API_KEY` env var) |
| `--base-url` | | Public URL for download links |
| `--db` | `webfetch.db` | SQLite database path |

| Endpoint | Auth | Description |
|----------|------|-------------|
| `POST /mcp` | Yes | JSON-RPC 2.0 handler |
| `GET /results/{id}` | Yes | Download oversized results |
| `POST /admin/tokens` | Yes | Create file access tokens |
| `GET /health` | No | Health check |

Register with Claude Code via HTTP transport:

```bash
claude mcp add --transport http webfetch-clean http://localhost:8080/mcp
```

### Docker Deployment

```bash
export WEBFETCH_API_KEY=your-secret-key
export BASE_URL=https://fetch.example.com
export SITE_ADDRESS=fetch.example.com

docker compose up -d
```

Includes Caddy reverse proxy with automatic TLS.

## Architecture

```
Input (URL or File) -> Fetch/Read -> Clean HTML -> Convert to Format -> Output
```

| File | Purpose |
|------|---------|
| `main.go` | Entry point, MCP protocol, CLI routing |
| `httpserver.go` | HTTP server, auth middleware, TempStore |
| `tokenstore.go` | SQLite-backed file access tokens |
| `fetcher.go` | HTTP client |
| `cleaner.go` | Multi-pass HTML cleaning pipeline |
| `converter.go` | HTML-to-Markdown conversion |
| `db/` | sqlc-generated database code |

### Dependencies

- [goquery](https://github.com/PuerkitoBio/goquery) - HTML parsing
- [html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) - Markdown conversion
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) - Pure Go SQLite (no CGO)

## Testing

```bash
# All tests
go test -v ./...

# With coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# With race detection
go test -v -race ./...
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, coding standards, and PR process.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- [goquery](https://github.com/PuerkitoBio/goquery) by Martin Angers
- [html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) by Johannes Kaufmann
- [MCP protocol](https://modelcontextprotocol.io/) by Anthropic
