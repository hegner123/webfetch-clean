# webfetch-clean CLI Tool Fixes

## Overview

Fix 18 issues identified in CLI tool review. Organized for multi-agent parallel development using HQ coordination.

## HQ Project Structure

**Project:** `webfetch-cli-fixes`
**Base Commit:** `8a14eb39ab60b50d2dc3efb8a69198c6f47a8b87`

### Parallel Execution Map

```
WAVE 1 (12 steps - no dependencies, can all run in parallel):
├── fetcher scope:
│   ├── Step 1: response-size-limit
│   ├── Step 2: url-scheme-validation
│   └── Step 3: timeout-default
├── main-cli scope:
│   ├── Step 6: version-flag
│   ├── Step 7: verbose-flag
│   └── Step 8: exit-codes
├── main-mcp scope:
│   ├── Step 10: scanner-buffer
│   └── Step 11: output-directory
├── main-files scope:
│   ├── Step 13: symlink-safety
│   ├── Step 14: windows-path-handling
│   └── Step 15: file-collision-prevention
└── quality scope:
    └── Step 18: go-mod-tidy

WAVE 2 (depends on Wave 1):
├── Step 4: redirect-tracking-and-context → depends on [1, 2, 3]
├── Step 9: help-text → depends on [6, 7, 8]
├── Step 12: signal-handling → depends on [10]
└── Step 16: remove-custom-string-funcs → depends on [14]

WAVE 3 (depends on Wave 2):
├── Step 5: fetcher-tests → depends on [4]
└── Step 17: rename-processURL → depends on [4, 13, 14, 15, 16]
```

### Step Details

| Step | Branch | Scope | Depends On | Description |
|------|--------|-------|------------|-------------|
| 1 | response-size-limit | fetcher | - | Add 50MB limit with io.LimitReader |
| 2 | url-scheme-validation | fetcher | - | Validate http/https only |
| 3 | timeout-default | fetcher | - | Default to 30s if timeout <= 0 |
| 4 | redirect-tracking-and-context | fetcher | 1,2,3 | Return finalURL, add context.Context |
| 5 | fetcher-tests | fetcher | 4 | Tests for all fetcher changes |
| 6 | version-flag | main-cli | - | Add --version flag |
| 7 | verbose-flag | main-cli | - | Add --verbose for progress |
| 8 | exit-codes | main-cli | - | Meaningful exit codes (2,3,4) |
| 9 | help-text | main-cli | 6,7,8 | Custom flag.Usage with examples |
| 10 | scanner-buffer | main-mcp | - | Increase to 10MB buffer |
| 11 | output-directory | main-mcp | - | Add output_dir parameter |
| 12 | signal-handling | main-mcp | 10 | Graceful SIGINT/SIGTERM |
| 13 | symlink-safety | main-files | - | Use os.Lstat, reject non-regular |
| 14 | windows-path-handling | main-files | - | Use filepath.Base/Ext |
| 15 | file-collision-prevention | main-files | - | Unique filenames with suffix |
| 16 | remove-custom-string-funcs | main-files | 14 | Use strings stdlib |
| 17 | rename-processURL | main-process | 4,13-16 | Rename to processInput |
| 18 | go-mod-tidy | quality | - | Fix indirect markers |

---

## Implementation Details

### Step 1: Response Size Limit (fetcher.go)
```go
const MaxResponseSize = 50 * 1024 * 1024 // 50MB

limitedReader := io.LimitReader(resp.Body, MaxResponseSize+1)
body, err := io.ReadAll(limitedReader)
if len(body) > MaxResponseSize {
    return "", fmt.Errorf("response too large: exceeds %d bytes", MaxResponseSize)
}
```

### Step 2: URL Scheme Validation (fetcher.go)
```go
if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
    return "", fmt.Errorf("invalid URL scheme: only http:// and https:// supported")
}
```

### Step 3: Timeout Default (fetcher.go)
```go
const DefaultTimeout = 30
if timeout <= 0 {
    timeout = DefaultTimeout
}
```

### Step 4: Redirect Tracking + Context (fetcher.go)
Combined changes to FetchURL signature:
```go
func FetchURL(ctx context.Context, url string, timeout int) (content, finalURL string, err error)
req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
// After request:
finalURL = resp.Request.URL.String()
```

### Step 6: Version Flag (main.go)
```go
const Version = "1.0.0"
flag.BoolVar(&showVersion, "version", false, "Print version and exit")
```

### Step 8: Exit Codes (main.go)
```go
const (
    ExitSuccess = 0
    ExitGeneral = 1
    ExitUsage   = 2
    ExitNetwork = 3
    ExitFileIO  = 4
)
```

### Step 10: Scanner Buffer (main.go)
```go
const MaxScannerBuffer = 10 * 1024 * 1024 // 10MB
scanner := bufio.NewScanner(os.Stdin)
scanner.Buffer(make([]byte, 0, MaxScannerBuffer), MaxScannerBuffer)
```

### Step 14: Windows Path Handling (main.go)
Replace manual slash parsing with:
```go
base := filepath.Base(filePath)
ext := filepath.Ext(base)
name := strings.TrimSuffix(base, ext) + "-cleaned"
```

### Step 16: Remove Custom String Functions (main.go)
Delete lines 530-569 and replace usages:
- `replaceAll` → `strings.ReplaceAll`
- `containsStr` → `strings.Contains`
- `trimHyphens` → `strings.Trim(s, "-")`

---

## Agent Commands

### Check available work:
```bash
# Using MCP tool
mcp__hq__get_available_steps project="webfetch-cli-fixes"
```

### Claim and start a step:
```bash
mcp__hq__claim_step project="webfetch-cli-fixes" agent_id="agent-1"
mcp__hq__start_step step_id=<id>
# Send heartbeats every 30-60s while working
mcp__hq__heartbeat step_id=<id> agent_id="agent-1"
```

### Complete a step:
```bash
mcp__hq__complete_step step_id=<id> commit_hash="abc123" files_modified='["fetcher.go"]'
```

---

## Verification

After each step:
1. `go build` - compiles
2. `go test -v ./...` - tests pass
3. `go test -race ./...` - no races
4. `golangci-lint run` - no lint errors

Final verification:
5. `echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | ./webfetch-clean`
6. `./webfetch-clean --cli --url https://example.com`

## Critical Files

- `main.go` - Entry point, CLI, MCP protocol
- `fetcher.go` - HTTP client, security fixes
- `fetcher_test.go` - Test patterns
- `go.mod` - Needs tidy
