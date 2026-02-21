# webfetch-clean - Getting Started

## Onboarding

Welcome to the webfetch-clean MCP tool project! This tool fetches web pages, removes clutter (ads, scripts, navigation), and outputs clean HTML or Markdown.

### Quick Start

The project is in the testing phase. Core implementation is complete:
- Binary built and installed: `/usr/local/bin/webfetch-clean`
- MCP server configured: `.mcp.json`
- Documentation complete: README.md, CONTRIBUTING.md, CLAUDE.md updated

Next steps: Comprehensive testing of CLI and MCP modes.

### Project Goal

Build an MCP tool that Claude Code can use to fetch web content with much cleaner formatting than the built-in WebFetch tool by:
- Removing `<script>`, `<style>`, `<nav>`, `<head>` tags
- Removing ad-related elements
- Removing navigation, sidebars, and clutter
- Converting to clean Markdown or HTML output

## Active Step

**Current Phase:** Phase 7 - Testing

**Current Step:** Comprehensive testing of CLI and MCP modes

**Status:** Ready to test

**Next Actions:**
1. Test CLI mode with multiple URLs (news sites, documentation, blogs)
2. Test MCP mode with JSON-RPC protocol
3. Verify both HTML and Markdown output formats
4. Test edge cases (timeouts, 404s, malformed HTML)
5. Verify MCP integration with Claude Code

**Implementation Status:**
- Binary built and installed to `/usr/local/bin` ✅
- All core components implemented (fetcher, cleaner, converter, main) ✅
- Documentation complete (README, CONTRIBUTING, CLAUDE.md) ✅
- MCP configuration created (.mcp.json) ✅

---

## Implementation Phases

### ✅ Phase 1: Project Setup (COMPLETE)
- [x] Create project directory
- [x] Write WEBFETCH-CLEAN-IMPLEMENTATION-PLAN.md
- [x] Write START.md
- [x] Initialize Go module
- [x] Add Go dependencies
- [x] Create `.gitignore`

### ✅ Phase 2: Core HTTP Fetcher (COMPLETE)
- [x] Create `fetcher.go`
- [x] Implement `FetchURL()` function
- [x] Add timeout handling
- [x] Add error handling for HTTP status codes
- [x] Test with various URLs

### ✅ Phase 3: HTML Cleaning Logic (COMPLETE)
- [x] Create `cleaner.go`
- [x] Implement `CleanHTML()` function
- [x] Add multi-pass cleaning pipeline
- [x] Add `preserveMainOnly` option
- [x] Add `removeImages` option
- [x] Test cleaning effectiveness

### ✅ Phase 4: Format Conversion (COMPLETE)
- [x] Create `converter.go`
- [x] Implement `ConvertToMarkdown()` function
- [x] Handle both HTML and Markdown output
- [x] Test conversion quality

### ✅ Phase 5: Main Entry Point (COMPLETE)
- [x] Create `main.go`
- [x] Copy MCP protocol from checkfor
- [x] Implement dual-mode routing (CLI vs MCP)
- [x] Add CLI flag parsing
- [x] Wire up pipeline: fetcher → cleaner → converter
- [x] Test MCP JSON-RPC methods

### ✅ Phase 6: Configuration (COMPLETE)
- [x] Create `.mcp.json`
- [x] Test MCP server configuration

### 🔄 Phase 7: Testing (IN PROGRESS)
- [ ] Test CLI mode with news sites
- [ ] Test CLI mode with documentation sites
- [ ] Test CLI mode with blogs
- [ ] Test MCP mode with JSON-RPC
- [ ] Test both output formats
- [ ] Test edge cases (timeouts, 404s, malformed HTML)

### ✅ Phase 8: Installation (COMPLETE)
- [x] Build binary
- [x] Install to `/usr/local/bin`
- [x] Verify CLI functionality
- [ ] Verify MCP integration with Claude Code

### ✅ Phase 9: Documentation (COMPLETE)
- [x] Update `/Users/home/.claude/CLAUDE.md`
- [x] Write README.md
- [x] Write CONTRIBUTING.md
- [x] Write LICENSE
- [x] Document all CLI flags
- [x] Add usage examples

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
- [x] Binary builds without errors
- [ ] CLI mode works with multiple URLs (needs testing)
- [ ] MCP mode responds to JSON-RPC correctly (needs testing)
- [ ] HTML output format works (needs testing)
- [ ] Markdown output format works (needs testing)
- [ ] Cleaning removes ads, scripts, nav effectively (needs testing)
- [ ] Semantic content is preserved (needs testing)
- [ ] Tool appears in Claude Code (needs verification)
- [ ] Claude automatically uses tool when appropriate (needs verification)

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

## Quick Start Reference

**Reference Files:**
- Implementation Plan: `WEBFETCH-CLEAN-IMPLEMENTATION-PLAN.md`
- MCP Protocol Pattern: `/Users/home/Documents/Code/Go_dev/checkfor/main.go`
- Configuration Pattern: `/Users/home/Documents/Code/Go_dev/checkfor/.mcp.json`

**Current Files:**
- `main.go` - Entry point and MCP protocol handler
- `fetcher.go` - HTTP client with timeout handling
- `cleaner.go` - Multi-pass HTML cleaning pipeline
- `converter.go` - HTML-to-Markdown conversion
- `.mcp.json` - MCP server configuration
- `go.mod` / `go.sum` - Go module dependencies

---

**Last Updated:** 2026-01-13 (Core implementation complete, testing phase)
**Next Review:** After Phase 7 (Testing) completion
