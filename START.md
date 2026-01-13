# webfetch-clean - Getting Started

## Onboarding

Welcome to the webfetch-clean MCP tool project! This tool fetches web pages, removes clutter (ads, scripts, navigation), and outputs clean HTML or Markdown.

### Quick Start

1. **Read the implementation plan**: See `WEBFETCH-CLEAN-IMPLEMENTATION-PLAN.md` for the complete technical design
2. **Follow the active step below** to track current progress
3. **Reference files**:
   - `/Users/home/Documents/Code/Go_dev/checkfor/main.go` - MCP protocol pattern to follow
   - `/Users/home/Documents/Code/Go_dev/checkfor/.mcp.json` - Configuration pattern

### Project Goal

Build an MCP tool that Claude Code can use to fetch web content with much cleaner formatting than the built-in WebFetch tool by:
- Removing `<script>`, `<style>`, `<nav>`, `<head>` tags
- Removing ad-related elements
- Removing navigation, sidebars, and clutter
- Converting to clean Markdown or HTML output

## Active Step

**Current Phase:** Phase 1 - Project Setup

**Current Step:** Initialize Go module and add dependencies

**Status:** Ready to execute

**Next Actions:**
1. Run `go mod init github.com/hegner123/webfetch-clean`
2. Run `go get github.com/PuerkitoBio/goquery`
3. Run `go get github.com/JohannesKaufmann/html-to-markdown`
4. Create `.gitignore` file

**Dependencies:**
- Go 1.23 or later
- Internet connection for fetching dependencies

---

## Implementation Phases

### ✅ Phase 1: Project Setup (IN PROGRESS)
- [x] Create project directory
- [x] Write WEBFETCH-CLEAN-IMPLEMENTATION-PLAN.md
- [x] Write START.md
- [ ] Initialize Go module
- [ ] Add Go dependencies
- [ ] Create `.gitignore`

### ⬜ Phase 2: Core HTTP Fetcher
- [ ] Create `fetcher.go`
- [ ] Implement `FetchURL()` function
- [ ] Add timeout handling
- [ ] Add error handling for HTTP status codes
- [ ] Test with various URLs

### ⬜ Phase 3: HTML Cleaning Logic
- [ ] Create `cleaner.go`
- [ ] Implement `CleanHTML()` function
- [ ] Add multi-pass cleaning pipeline
- [ ] Add `preserveMainOnly` option
- [ ] Add `removeImages` option
- [ ] Test cleaning effectiveness

### ⬜ Phase 4: Format Conversion
- [ ] Create `converter.go`
- [ ] Implement `ConvertToMarkdown()` function
- [ ] Handle both HTML and Markdown output
- [ ] Test conversion quality

### ⬜ Phase 5: Main Entry Point
- [ ] Create `main.go`
- [ ] Copy MCP protocol from checkfor
- [ ] Implement dual-mode routing (CLI vs MCP)
- [ ] Add CLI flag parsing
- [ ] Wire up pipeline: fetcher → cleaner → converter
- [ ] Test MCP JSON-RPC methods

### ⬜ Phase 6: Configuration
- [ ] Create `.mcp.json`
- [ ] Test MCP server configuration

### ⬜ Phase 7: Testing
- [ ] Test CLI mode with news sites
- [ ] Test CLI mode with documentation sites
- [ ] Test CLI mode with blogs
- [ ] Test MCP mode with JSON-RPC
- [ ] Test both output formats
- [ ] Test edge cases (timeouts, 404s, malformed HTML)

### ⬜ Phase 8: Installation
- [ ] Build binary
- [ ] Install to `/usr/local/bin`
- [ ] Verify CLI functionality
- [ ] Verify MCP integration with Claude Code

### ⬜ Phase 9: Documentation
- [ ] Update `/Users/home/.claude/CLAUDE.md`
- [ ] Write README.md
- [ ] Document all CLI flags
- [ ] Add usage examples

---

## Development Commands

### Setup
```bash
cd /Users/home/Documents/Code/Go_dev/webfetch-clean
go mod init github.com/hegner123/webfetch-clean
go get github.com/PuerkitoBio/goquery
go get github.com/JohannesKaufmann/html-to-markdown
```

### Build
```bash
go build -o webfetch-clean
```

### Test CLI Mode
```bash
./webfetch-clean --url https://example.com --format markdown
./webfetch-clean --url https://news.ycombinator.com --format html
```

### Test MCP Mode
```bash
# Initialize
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | ./webfetch-clean --mcp

# List tools
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | ./webfetch-clean --mcp

# Call tool
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"webfetch_clean","arguments":{"url":"https://example.com"}}}' | ./webfetch-clean --mcp
```

### Install
```bash
sudo cp webfetch-clean /usr/local/bin/
webfetch-clean --help
```

---

## Success Criteria

Before marking the project complete, verify:

- [x] Project structure created
- [x] Implementation plan documented
- [ ] Binary builds without errors
- [ ] CLI mode works with multiple URLs
- [ ] MCP mode responds to JSON-RPC correctly
- [ ] HTML output format works
- [ ] Markdown output format works
- [ ] Cleaning removes ads, scripts, nav effectively
- [ ] Semantic content is preserved
- [ ] Tool appears in Claude Code
- [ ] Claude automatically uses tool when appropriate

---

## Troubleshooting

### Go Module Issues
```bash
go mod tidy
go mod download
```

### Build Issues
```bash
go clean
go build -v -o webfetch-clean
```

### Dependency Issues
```bash
go get -u github.com/PuerkitoBio/goquery
go get -u github.com/JohannesKaufmann/html-to-markdown
```

---

## Resources

- **goquery Documentation**: https://github.com/PuerkitoBio/goquery
- **html-to-markdown Documentation**: https://github.com/JohannesKaufmann/html-to-markdown
- **MCP Protocol Spec**: See Claude Code documentation
- **Reference Implementation**: `/Users/home/Documents/Code/Go_dev/checkfor/main.go`

---

**Last Updated:** Project initialization
**Next Review:** After Phase 1 completion
