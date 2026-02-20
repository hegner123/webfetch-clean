# Plan: HTTP Server Mode with Token-Based File Access

## Context

webfetch-clean currently runs as a stdio MCP server or CLI tool. To allow remote team members and an AI orchestrator migration tool to access it without local installation, we need an HTTP server mode.

**Audience:** Remote team members and AI orchestrator migration tool.
**Deployment:** Containerized behind Caddy reverse proxy in Docker Compose.
**Security model:** API key auth, token-gated file access, no raw file paths over HTTP.

Key design challenges:
1. Over-limit responses write to local files — doesn't work for remote callers. HTTP server stores oversized content in memory with TTL and gives agents retrieval options.
2. The `file` parameter reads arbitrary local paths — remote callers must not have that access. File access uses single-use UUID tokens stored in SQLite.
3. Modes are fully separated: `--http` starts HTTP-only, no flag starts stdio MCP-only.

## Approach

Refactor `processInput()` to remove I/O side effects (pure pipeline). Add HTTP server mode behind `--http` flag. File access in HTTP mode uses SQLite-backed tokens (sqlc for type-safe queries). Temp store for oversized responses is in-memory with bounded capacity.

## Files Modified

- **main.go** — Config fields, processInput refactor, flag parsing, mode routing
- **limit_test.go** — Update 2 tests that assert on file-write behavior (now assert on OverLimit fields)

## New Files

- **sqlc.yaml** — sqlc configuration
- **sqlc/schema.sql** — file_tokens table DDL
- **sqlc/query.sql** — token CRUD queries
- **db/** — sqlc-generated Go code (package `db`)
- **httpserver.go** — HTTP server, handlers, temp store, cleanup goroutine, auth middleware
- **tokenstore.go** — Token management layer over sqlc-generated code (create, validate, consume)
- **httpserver_test.go** — HTTP endpoint tests, temp store tests
- **tokenstore_test.go** — Token lifecycle tests

## New Dependencies

- `modernc.org/sqlite` — Pure Go SQLite driver (no CGO, container-friendly)
- `sqlc` — Build-time code generation (not a runtime dependency)

## Implementation Steps

### Step 1: Add fields to CleanResult and Config (main.go)

CleanResult — add 3 fields:
```go
TokenCount int    `json:"token_count,omitempty"`
OverLimit  bool   `json:"-"`
RawContent string `json:"-"`
```

Config — add 4 fields:
```go
HTTPAddr string
APIKey   string
BaseURL  string
DBPath   string
```

### Step 2: Refactor processInput

In `processInput()`, find the block starting with the comment `// Step 4: Check token limit and write to file if exceeded` and replace everything from that comment through the closing `return result` of that branch (the `result.Content = fmt.Sprintf(...)` block). Replace with:

```go
// Step 4: Check token limit — signal over-limit to caller, no I/O here
// result.Content remains empty; callers handle the over-limit case
tokenCount := len(output) / 3
if config.MaxTokens > 0 && tokenCount > config.MaxTokens {
    result.OverLimit = true
    result.TokenCount = tokenCount
    result.RawContent = output
    return result
}
```

processInput becomes a pure pipeline with no I/O side effects at the limit-check step.

### Step 3: Move file-write into stdio callers

The stdio `handleToolsCall` in main.go and `runCLI` in main.go each get the file-write logic that was removed from `processInput`. The HTTP server in httpserver.go has its own `handleHTTPToolsCall` that is completely separate — it does tempStore on OverLimit instead (see Step 7). The stdio and HTTP handlers share `processInput()` but diverge on how they handle results.

**handleToolsCall** — After `result := processInput(config)`, add the over-limit file-write block:
```go
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
    result.RawContent = "" // free memory
}
```

**runCLI** — After `result := processInput(config)` and the error check, add the same over-limit block before the existing output handling:
```go
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
    result.RawContent = "" // free memory
}
```

### Step 4: Update limit_test.go

Two tests currently assert result.Content contains "Output exceeded limit" and "Content written to file:". After refactoring, processInput sets result.OverLimit = true and result.RawContent instead.

**TestProcessInput_OutputLimit_Exceeded** — Assert:
- `result.OverLimit == true`
- `result.TokenCount > 0`
- `result.RawContent` contains the cleaned content
- `result.Content == ""` (no file-write message)

**TestProcessInput_OutputLimit_URLFilename** — Assert:
- `result.OverLimit == true`
- `result.RawContent != ""`

### Step 5: Create sqlc configuration and SQL files

**sqlc.yaml:**
```yaml
version: "2"
sql:
  - engine: "sqlite"
    queries: "sqlc/query.sql"
    schema: "sqlc/schema.sql"
    gen:
      go:
        package: "db"
        out: "db"
```

**sqlc/schema.sql:**
```sql
-- expires stores Unix epoch seconds (integer) for portable time comparison.
-- All time comparisons use unixepoch('now') to avoid ISO 8601 format mismatches
-- between Go's time.Time serialization and SQLite's datetime() function.
CREATE TABLE IF NOT EXISTS file_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token TEXT UNIQUE NOT NULL,
    file TEXT NOT NULL,
    expires INTEGER NOT NULL,
    consumed BOOLEAN NOT NULL DEFAULT FALSE
);
```

**sqlc/query.sql:**
```sql
-- name: CreateToken :one
INSERT INTO file_tokens (token, file, expires, consumed)
VALUES (?, ?, ?, FALSE)
RETURNING *;

-- name: RedeemToken :one
-- Atomic single-use redemption: marks consumed and returns file path in one statement.
-- Prevents TOCTOU race where two concurrent requests could both redeem the same token.
-- Returns no rows if token is invalid, expired, or already consumed.
UPDATE file_tokens SET consumed = TRUE
WHERE token = ? AND consumed = FALSE AND expires > unixepoch('now')
RETURNING file;

-- name: DeleteExpired :exec
DELETE FROM file_tokens WHERE expires < unixepoch('now') OR consumed = TRUE;

-- name: CountActiveTokens :one
SELECT COUNT(*) FROM file_tokens WHERE consumed = FALSE AND expires > unixepoch('now');
```

Run `sqlc generate` to produce the `db/` package.

### Step 6: Create tokenstore.go

Token management layer wrapping the sqlc-generated code:

```go
type TokenStore struct {
    queries *db.Queries
    sqlDB   *sql.DB
}
```

Functions:
- `NewTokenStore(dbPath string) (*TokenStore, error)` — Opens SQLite, enables WAL mode and busy timeout, runs schema migration, returns store. Connection init runs:
  ```sql
  PRAGMA journal_mode=WAL;
  PRAGMA busy_timeout=5000;
  ```
  WAL mode allows concurrent reads during writes. Busy timeout causes SQLite to retry for 5 seconds instead of returning SQLITE_BUSY immediately.
- `(ts *TokenStore) CreateFileToken(filePath string, ttl time.Duration) (string, error)` — Generates UUID, stores `expires` as Unix epoch seconds (`time.Now().Add(ttl).Unix()`), inserts row, returns token string. Validates file exists and is readable before creating token.
- `(ts *TokenStore) RedeemToken(token string) (string, error)` — Atomic single-statement redemption via `UPDATE ... WHERE consumed = FALSE AND expires > unixepoch('now') RETURNING file`. Returns file path if exactly one row affected. Returns error if no rows affected (not found, expired, or already consumed). No TOCTOU race — the UPDATE and the validity check are a single atomic operation.
- `(ts *TokenStore) Cleanup() error` — Deletes expired and consumed rows
- `(ts *TokenStore) Close() error` — Closes SQLite connection

`generateUUID() string` — RFC 4122 v4 using crypto/rand, no new deps. This is a package-level function in tokenstore.go. Since both tokenstore.go and httpserver.go are in package main, it is automatically available to both without import.

Schema migration: `NewTokenStore` executes the `schema.sql` contents using `db.ExecContext` with `CREATE TABLE IF NOT EXISTS` semantics. No schema versioning needed for v1.

### Step 7: Create httpserver.go

**Server struct** (no package-level globals):
```go
type HTTPServer struct {
    config     Config
    tempStore  *TempStore
    tokenStore *TokenStore
    mux        *http.ServeMux
}
```

**TempStore type** (bounded in-memory store):
```go
type TempEntry struct {
    Content   string
    Format    string
    URL       string
    CreatedAt time.Time
}

type TempStore struct {
    mu         sync.RWMutex
    entries    map[string]TempEntry
    maxEntries int
}
```

Constants:
```go
const tempStoreTTL = 60 * time.Second
```
Reference this constant in the cleanup goroutine, the instructions text, and tests. Do not hardcode `60` in multiple places.

Methods: `Store(id string, entry TempEntry) error` (capacity check is inside the mutex lock — acquire lock, check len >= maxEntries, if full return error, otherwise insert; returns error if at capacity), `LoadAndDelete(id string) (TempEntry, bool)`, `Cleanup()`, `Len() int`

**Max entries:** 100 (configurable). At 300KB+ per entry, this caps memory at ~30MB worst case.

**runHTTPServer(config Config):**
1. Initialize TokenStore (opens SQLite at config.DBPath)
2. Initialize TempStore with max entries
3. Build HTTPServer struct with mux
4. Register routes on mux
5. Start temp store cleanup goroutine (tick every 10s, delete entries older than 60s)
6. Start token store cleanup goroutine (tick every 5min, delete expired/consumed rows)
7. Start HTTP server with timeouts
8. Graceful shutdown on SIGINT/SIGTERM (first signal: Shutdown with 30s context; second signal: os.Exit(1))

**Timeouts:**
- Read: 30s
- Write: 120s (fetching remote URLs can be slow)
- Idle: 60s
- Request body limit: 10MB via http.MaxBytesReader (matches existing MaxScannerBuffer)

**Default bind address:** 127.0.0.1 if only port specified. `--http :8080` binds to `127.0.0.1:8080`. `--http 0.0.0.0:8080` required for all-interface binding. Normalization code in `runHTTPServer`:
```go
addr := config.HTTPAddr
if strings.HasPrefix(addr, ":") {
    addr = "127.0.0.1" + addr
}
```

**Routes:**
- `POST /mcp` — JSON-RPC handler (API key required)
- `GET /results/{id}` — Download oversized content (API key required)
- `POST /admin/tokens` — Create file token (API key required)
- `GET /health` — Health check (no auth)

**Auth middleware:**
- Compare `X-API-Key` header against config.APIKey using `crypto/subtle.ConstantTimeCompare` (prevents timing attacks)
- Return 401 if missing or mismatched
- Skip auth for GET /health

**POST /mcp handler:**
- http.MaxBytesReader on request body (10MB)
- Parse JSON-RPC, route to method handlers: `initialize`, `notifications/initialized`, `tools/list`, `tools/call`. Method routing logic is identical to the stdio handler except `tools/call` uses the HTTP-specific schema and over-limit behavior.
- Method handlers return (JSONRPCResponse, error) — no stdout writes
- Response written to http.ResponseWriter as JSON
- If a `file` parameter is present in `tools/call` arguments, actively reject with error: `sendError(req.ID, ErrInvalidParams, "'file' parameter is not available in HTTP mode. Use 'file_token' instead.")`. Silent ignore would cause agents to retry indefinitely.

**HTTP tool schema** (different from MCP stdio schema):
- Has `url`, `output_format`, `mode`, `preserve_main_only`, `remove_images`, `strip_links`, `use_browser`, `timeout`, `max_tokens`
- Has `file_token` instead of `file`: `{Type: "string", Description: "Single-use token for server-side file access (replaces file parameter)"}`
- Has `result_id`: `{Type: "string", Description: "ID of a stored over-limit result. Use with override=true to retrieve full content."}`
- Has `override`: `{Type: "boolean", Description: "When true with result_id, returns the full stored content.", Default: false}`
- Does NOT have `file` or `output_directory`

**tools/call handler — file_token path:**
1. Extract `file_token` from arguments
2. Call tokenStore.RedeemToken(file_token) — returns file path
3. Set config.File to the returned path
4. Continue through normal processInput pipeline

**tools/call handler — result_id + override path:**
1. At top of handler, before URL/file validation
2. Look up entry in tempStore.LoadAndDelete(result_id)
3. Return content directly, entry is removed
4. 404-equivalent error if not found or expired

**tools/call handler — over-limit path:**
1. processInput returns result.OverLimit = true
2. Store result.RawContent in tempStore with UUID
3. If tempStore.Store returns capacity error, fall back to truncation with message
4. Build response with both structured `result_id` field and human-readable instructions text:

The over-limit CleanResult includes `result_id` as a structured JSON field for programmatic extraction:
```go
type OverLimitResult struct {
    ResultID   string `json:"result_id"`
    TokenCount int    `json:"token_count"`
    Limit      int    `json:"limit"`
    Message    string `json:"message"`
}
```

The `Message` field contains the human-readable instructions:
```
Fetched {url} — {tokenCount} tokens (limit: {maxTokens}).
Result stored for 60 seconds (ID: {uuid}).

Options:
1. Override limit: call webfetch_clean with {"result_id": "{uuid}", "override": true}
2. Download: GET {baseURL}/results/{uuid}
3. Do nothing — result expires in 60 seconds.
```

This allows tests and agents to extract `result_id` from the JSON structure directly (`response.result.content[0].text` parsed as JSON → `.result_id`) rather than parsing prose.

**GET /results/{id} handler:**
- r.PathValue("id") to extract UUID
- tempStore.LoadAndDelete(id)
- Set Content-Type: text/markdown or text/html based on entry.Format
- Content-Disposition: attachment with filename
- 404 if not found or expired

**POST /admin/tokens handler:**
- Parse JSON body: `{"file": "/data/path.html", "expires_minutes": 60}`
- `expires_minutes` must be a positive integer. Default: 60. Maximum: 1440 (24 hours). Return 400 if out of range or non-positive.
- Call tokenStore.CreateFileToken(file, duration)
- Return JSON: `{"token": "uuid-here", "expires": "2024-01-01T00:00:00Z"}`
- 400 if file doesn't exist or isn't readable

**Cleanup goroutines:**
- Temp store: time.NewTicker(10s), context-cancellable, deletes entries older than 60s
- Token store: time.NewTicker(5min), context-cancellable, deletes expired/consumed rows

### Step 8: Add flags to parseFlags (main.go)

```go
flag.StringVar(&config.HTTPAddr, "http", "", "Start HTTP server on address (e.g., :8080)")
flag.StringVar(&config.APIKey, "api-key", "", "API key for HTTP mode (or WEBFETCH_API_KEY env)")
flag.StringVar(&config.BaseURL, "base-url", "", "Public base URL for download links (e.g., https://fetch.example.com)")
flag.StringVar(&config.DBPath, "db", "webfetch.db", "SQLite database path for file tokens (HTTP mode)")
```

API key resolution: flag value takes precedence, falls back to WEBFETCH_API_KEY env var. HTTP mode requires a non-empty API key — exits with error if missing.

BaseURL normalization: strip trailing slashes at parse time to prevent double-slash URLs in download links (e.g., `https://example.com/` + `/results/uuid` → `https://example.com//results/uuid`).
```go
config.BaseURL = strings.TrimRight(config.BaseURL, "/")
```

Mode routing in main():
```go
if config.HTTPAddr != "" {
    runHTTPServer(config)
    return
}
```

Update printUsage with HTTP mode section.

### Step 9: Create httpserver_test.go

Using httptest.NewServer wrapping the HTTPServer's mux. Each test creates its own HTTPServer with in-memory SQLite (`:memory:`) and fresh TempStore.

**TempStore tests:**
- TestTempStore_StoreAndLoadAndDelete — store, load+delete, verify, verify second load returns false
- TestTempStore_Cleanup — store entry with old timestamp, cleanup, verify removed
- TestTempStore_Cleanup_PreservesRecent — store recent entry, cleanup, verify survives
- TestTempStore_Store_AtCapacity — fill store to max, verify next Store returns error
- TestTempStore_Concurrent — goroutines doing Store + LoadAndDelete + Cleanup simultaneously

**UUID tests:**
- TestGenerateUUID_Format — verify 8-4-4-4-12 hex format
- TestGenerateUUID_Unique — generate 1000, verify no dupes

**Auth tests:**
- TestHTTP_Auth_Missing — request without API key, verify 401
- TestHTTP_Auth_Wrong — request with wrong API key, verify 401
- TestHTTP_Auth_Correct — request with correct API key, verify 200
- TestHTTP_Health_NoAuth — GET /health without API key, verify 200

**Endpoint tests:**
- TestHTTP_Initialize — POST /mcp with initialize, verify response
- TestHTTP_ToolsList — verify file_token, result_id, override in schema; verify no file parameter
- TestHTTP_ToolsCall_URL_UnderLimit — serve small HTML via httptest, verify content returned inline
- TestHTTP_ToolsCall_URL_OverLimit — serve large HTML with small max_tokens, verify instructions with UUID and options text
- TestHTTP_ToolsCall_Override — trigger over-limit, extract result_id, call again with override=true, verify full content returned
- TestHTTP_ToolsCall_Override_Expired — trigger over-limit, manually delete entry, try override, verify error
- TestHTTP_ResultDownload — trigger over-limit, GET /results/{id}, verify content returned and entry deleted
- TestHTTP_ResultDownload_NotFound — GET nonexistent ID, verify 404
- TestHTTP_ToolsCall_FileToken — create token via admin endpoint, use file_token in tools/call, verify file content returned
- TestHTTP_ToolsCall_FileToken_Consumed — use same token twice, verify second call fails
- TestHTTP_ToolsCall_FileToken_Expired — create token, manually expire it, verify call fails
- TestHTTP_ToolsCall_RawFile_Rejected — send `file` parameter (not file_token), verify error

**Admin tests:**
- TestHTTP_AdminTokens_Create — POST /admin/tokens, verify token returned
- TestHTTP_AdminTokens_FileNotFound — POST with nonexistent file, verify 400
- TestHTTP_AdminTokens_CustomExpiry — POST with expires_minutes, verify expiry

**Request limits:**
- TestHTTP_RequestBodyLimit — POST body > 10MB, verify 413 or error

### Step 10: Create tokenstore_test.go

Using in-memory SQLite (`:memory:`) for each test.

- TestTokenStore_CreateAndRedeem — create token, redeem, verify file path returned
- TestTokenStore_RedeemConsumed — create, redeem, redeem again, verify error on second
- TestTokenStore_RedeemExpired — create with short TTL, sleep past expiry, verify error
- TestTokenStore_RedeemNotFound — redeem nonexistent token, verify error
- TestTokenStore_Cleanup — create expired + consumed tokens, cleanup, verify removed
- TestTokenStore_CreateValidatesFile — create token for nonexistent file, verify error
- TestTokenStore_Close — create store, close, verify operations fail after close

### Step 11: Update CLAUDE.md

Add HTTP mode section documenting:
- `--http` flag and address binding (default 127.0.0.1)
- `--api-key` / WEBFETCH_API_KEY
- `--base-url` for Caddy proxy
- `--db` for SQLite path
- Endpoints: POST /mcp, GET /results/{id}, POST /admin/tokens, GET /health
- File token workflow
- Temp store behavior (60s TTL, 100 entry cap)
- HTTP tool schema differences from MCP schema

### Step 12: Docker and Caddy (follow-up, separate plan)

Not in scope for this plan, but noted for follow-up:
- Dockerfile (multi-stage build, scratch or distroless base)
- docker-compose.yml with Caddy and webfetch-clean services
- Caddyfile with reverse proxy, TLS, API key forwarding
- Shared volume for SQLite DB
- Health check configuration
- **SSRF mitigation:** Docker network configuration must restrict the webfetch-clean container's outbound access. The `url` parameter in `tools/call` allows fetching arbitrary URLs — without network-level restrictions, a remote caller could target internal services (`http://169.254.169.254/`, Docker-internal hostnames). Use a dedicated Docker network with no access to the host network or cloud metadata endpoints. Document this as a deployment requirement in the Compose setup.
- **Headless browser (`use_browser`) container requirements:** The HTTP schema includes `use_browser`, which launches headless Chromium via go-rod. The container image must include Chromium. Options:
  - `ghcr.io/go-rod/rod` — rod's official image (recommended, go-rod auto-detects container environment)
  - `chromedp/headless-shell` — minimal headless Chrome (~200MB)
  - DIY: `FROM golang:1.23-bookworm` + `apt-get install -y chromium`
  - Key container flags: `--no-sandbox`, `--disable-gpu`, `--disable-dev-shm-usage` (or mount `--shm-size=2g`). go-rod handles most flags automatically in container environments.
  - Resource requirements per concurrent browser tab: 512MB-2GB RAM, 1-2 vCPU, 1-5GB disk. Scale: 1 concurrent browser = 2 vCPU / 2GB min; 5 concurrent = 4 vCPU / 8GB; 10+ = consider a browser pool.

## sqlc Build Step

After creating the SQL files (Step 5), run:
```bash
sqlc generate
```

This produces the `db/` package with type-safe Go code. The generated code is committed to the repo (no build-time sqlc dependency for consumers).

## Verification

1. `sqlc generate` — produces db/ package without errors
2. `go build` — compiles cleanly with new dependencies
3. `go test -v ./...` — all existing tests pass (with limit_test.go updates)
4. `go test -v -run TestToken` — token store tests pass
5. `go test -v -run TestHTTP` — HTTP server tests pass
6. `go test -v -race ./...` — no races in concurrent temp store or token store access
7. Manual: `webfetch-clean --http :8080 --api-key test123 --base-url http://localhost:8080`
   - `curl localhost:8080/health` — 200
   - `curl -H "X-API-Key: test123" -X POST localhost:8080/mcp -d '{"jsonrpc":"2.0","id":1,"method":"initialize"}'`
   - `curl -H "X-API-Key: test123" -X POST localhost:8080/admin/tokens -d '{"file":"test.html"}'`
   - Use returned token in tools/call with file_token parameter
8. Verify stdio MCP mode unchanged: `echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | webfetch-clean`
