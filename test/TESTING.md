# Testing Guide

## Running Tests

```bash
# All tests
go test -v ./...

# Short mode (skips browser tests requiring Chromium)
go test -v -short ./...

# With coverage
go test -v -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -html=coverage.out   # HTML report
go tool cover -func=coverage.out   # Per-function breakdown

# With race detection
go test -v -race ./...

# Specific test
go test -v -run TestCleanHTML_RemoveAds
```

## Test Files

| File | Scope |
|------|-------|
| `browser_test.go` | Headless browser fetching (go-rod) |
| `cleaner_test.go` | HTML cleaning pipeline, scrape mode |
| `converter_test.go` | HTML/Markdown format conversion |
| `fetcher_test.go` | HTTP client, timeouts, redirects, scheme validation, size limits |
| `httpserver_test.go` | HTTP server endpoints, auth, TempStore, file tokens, over-limit handling |
| `integration_test.go` | Full pipeline (fetch/read, clean, convert) |
| `limit_test.go` | Output token limit and filename generation |
| `reader_test.go` | Local file reading, permissions, edge cases |
| `tokenstore_test.go` | SQLite token lifecycle (create, redeem, expire, cleanup) |
| `unique_filename_test.go` | Collision-safe filename generation |

## Coverage

Overall: ~52% of statements.

| Component | Coverage | Notes |
|-----------|----------|-------|
| `cleaner.go` | 90% | HTML cleaning pipeline |
| `fetcher.go` | 91% | HTTP client |
| `converter.go` | 86-100% | Format conversion |
| `reader.go` | 93% | Local file reading |
| `browser.go` | 79% | Headless browser (some tests require Chromium) |
| `httpserver.go` | 66-100% | HTTP endpoints and TempStore |
| `tokenstore.go` | 46-100% | SQLite token management |
| `limit.go` | 80-89% | Output limits and filenames |
| `main.go` | 0-70% | CLI/MCP entry points (processInput: 70%, handlers: 0%) |
| `db/` | 0% | sqlc-generated code (no tests, exercised via tokenstore) |

## CI/CD

GitHub Actions workflow: `.github/workflows/test.yml`

**Test matrix:**
- Platforms: Ubuntu, macOS, Windows
- Go versions: 1.25, 1.26

**Jobs:**
1. **test** - Run tests with `-short -race -coverprofile -covermode=atomic`
2. **build** - Build binary on all platforms, verify execution
3. **release** - Cross-compile and publish on version tags

Coverage uploaded to Codecov on ubuntu-latest with Go 1.26.

## Known Limitations

### Ad Detection

The ad detection is intentionally aggressive. Patterns like `-ad-`, `_ad_` may match legitimate class names (e.g., `thread-card`, `reader-mode`). This is acceptable because scrape mode (`--mode scrape`) bypasses all ad detection for cases where preservation matters.

### Browser Tests

Tests tagged with `skipping browser test in short mode` require a Chromium installation. Run without `-short` to execute them. CI uses `-short` to avoid browser dependency.

### Windows

- `TestReadFile_PermissionDenied` is skipped (chmod doesn't remove read access on Windows)
- File path tests use `json.Marshal` to handle backslash escaping in JSON payloads
